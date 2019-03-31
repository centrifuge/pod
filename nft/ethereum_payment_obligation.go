package nft

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/utils/byteutils"

	"github.com/centrifuge/go-centrifuge/utils/stringutils"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("nft")

const (
	// ErrNFTMinted error for NFT already minted for registry
	ErrNFTMinted = errors.Error("NFT already minted")
)

// Config is the config interface for nft package
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
}

// ethereumPaymentObligation handles all interactions related to minting of NFTs for payment obligations on Ethereum
type ethereumPaymentObligation struct {
	cfg             Config
	identityService identity.ServiceDID
	ethClient       ethereum.Client
	queue           queue.TaskQueuer
	docSrv          documents.Service
	bindContract    func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error)
	txManager       transactions.Manager
	blockHeightFunc func() (height uint64, err error)
}

// newEthereumPaymentObligation creates ethereumPaymentObligation given the parameters
func newEthereumPaymentObligation(
	cfg Config,
	identityService identity.ServiceDID,
	ethClient ethereum.Client,
	queue queue.TaskQueuer,
	docSrv documents.Service,
	bindContract func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error),
	txManager transactions.Manager,
	blockHeightFunc func() (uint64, error)) *ethereumPaymentObligation {
	return &ethereumPaymentObligation{
		cfg:             cfg,
		identityService: identityService,
		ethClient:       ethClient,
		bindContract:    bindContract,
		queue:           queue,
		docSrv:          docSrv,
		txManager:       txManager,
		blockHeightFunc: blockHeightFunc,
	}
}

func (s *ethereumPaymentObligation) filterMintProofs(docProof *documents.DocumentProof) *documents.DocumentProof {
	// Compact properties
	var nonFilteredProofsLiteral = [][]byte{append(documents.CompactProperties(documents.DRTreePrefix), documents.CompactProperties(documents.SigningRootField)...)}
	// Byte array Regex - (signatureTreePrefix + signatureProp) + Index[up to 104 characters (52bytes*2)] + Signature key
	m0 := append(documents.CompactProperties(documents.SignaturesTreePrefix), []byte{0, 0, 0, 1}...)
	var nonFilteredProofsMatch = []string{fmt.Sprintf("%s(.{104})%s", hex.EncodeToString(m0), hex.EncodeToString([]byte{0, 0, 0, 4}))}

	for i, p := range docProof.FieldProofs {
		if !byteutils.ContainsBytesInSlice(nonFilteredProofsLiteral, p.GetCompactName()) && !stringutils.ContainsBytesMatchInSlice(nonFilteredProofsMatch, p.GetCompactName()) {
			docProof.FieldProofs[i].SortedHashes = docProof.FieldProofs[i].SortedHashes[:len(docProof.FieldProofs[i].SortedHashes)-1]
		}
	}
	return docProof
}

func (s *ethereumPaymentObligation) prepareMintRequest(ctx context.Context, tokenID TokenID, cid identity.DID, req MintNFTRequest) (mreq MintRequest, err error) {
	docProofs, err := s.docSrv.CreateProofs(ctx, req.DocumentID, req.ProofFields)
	if err != nil {
		return mreq, err
	}

	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return mreq, err
	}

	pfs, err := model.CreateNFTProofs(cid,
		req.RegistryAddress,
		tokenID[:],
		req.SubmitTokenProof,
		req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof)
	if err != nil {
		return mreq, err
	}

	docProofs.FieldProofs = append(docProofs.FieldProofs, pfs...)
	docProofs = s.filterMintProofs(docProofs)

	anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
	if err != nil {
		return mreq, err
	}

	nextAnchorID, err := anchors.ToAnchorID(model.NextVersion())
	if err != nil {
		return mreq, err
	}

	proof, _ := documents.ConvertDocProofToClientFormat(&documents.DocumentProof{DocumentID: model.ID(), VersionID: anchorID[:], FieldProofs: docProofs.FieldProofs})
	log.Debug(json.MarshalIndent(proof, "", "  "))

	requestData, err := NewMintRequest(tokenID, req.DepositAddress, anchorID, nextAnchorID, docProofs.FieldProofs)
	if err != nil {
		return mreq, err
	}

	return requestData, nil

}

// GetRequiredInvoiceUnpaidProofFields returns required proof fields for an unpaid invoice mint
func (s *ethereumPaymentObligation) GetRequiredInvoiceUnpaidProofFields(ctx context.Context) ([]string, error) {
	var proofFields []string

	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}
	accDIDBytes, err := acc.GetIdentityID()
	if err != nil {
		return nil, err
	}
	keys, err := acc.GetKeys()
	if err != nil {
		return nil, err
	}

	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signerID := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerID)
	proofFields = []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due", "invoice.sender", "invoice.status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"}

	return proofFields, nil
}

// MintNFT mints an NFT
func (s *ethereumPaymentObligation) MintNFT(ctx context.Context, req MintNFTRequest) (*MintNFTResponse, chan bool, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	if !req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof {
		return nil, nil, errors.New("enable grant_nft_access to generate Read Access Proof")
	}

	tokenID := NewTokenID()
	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return nil, nil, err
	}

	// check if the nft is successfully minted already
	if model.IsNFTMinted(s, req.RegistryAddress) {
		return nil, nil, errors.NewTypedError(ErrNFTMinted, errors.New("registry %v", req.RegistryAddress.String()))
	}

	didBytes, err := tc.GetIdentityID()
	if err != nil {
		return nil, nil, err
	}

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, nil, err
	}
	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), did, transactions.NilTxID(), "Minting NFT",
		s.minter(ctx, tokenID, model, req))
	if err != nil {
		return nil, nil, err
	}

	return &MintNFTResponse{
		TransactionID: txID.String(),
		TokenID:       tokenID.String(),
	}, done, nil
}

func (s *ethereumPaymentObligation) minter(ctx context.Context, tokenID TokenID, model documents.Model, req MintNFTRequest) func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {
		err := model.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:])
		if err != nil {
			errOut <- err
			return
		}

		txctx := contextutil.WithTX(ctx, txID)
		_, _, done, err := s.docSrv.Update(txctx, model)
		if err != nil {
			errOut <- err
			return
		}

		isDone := <-done
		if !isDone {
			// some problem occurred in a child task
			errOut <- errors.New("update document failed for document %s and transaction %s", hexutil.Encode(req.DocumentID), txID)
			return
		}

		requestData, err := s.prepareMintRequest(txctx, tokenID, accountID, req)
		if err != nil {
			errOut <- errors.New("failed to prepare mint request: %v", err)
			return
		}

		// to common.Address, tokenId *big.Int, tokenURI string, anchorId *big.Int, properties [][]byte, values [][]byte, salts [][32]byte, proofs [][][32]byte
		utxID, done, err := s.identityService.Execute(ctx, req.RegistryAddress, EthereumPaymentObligationContractABI, "mint", requestData.To, requestData.TokenID, requestData.AnchorID, requestData.Props, requestData.Values, requestData.Salts, requestData.Proofs)
		if err != nil {
			errOut <- err
			return
		}
		log.Infof("Sent off ethTX to mint [tokenID: %s, anchor: %x, nextAnchor: %s, registry: %s] to payment obligation contract.",
			requestData.TokenID, requestData.AnchorID, hexutil.Encode(requestData.NextAnchorID.Bytes()), requestData.To.String())

		log.Debugf("To: %s", requestData.To.String())
		log.Debugf("TokenID: %s", hexutil.Encode(requestData.TokenID.Bytes()))
		log.Debugf("TokenURI: %s", requestData.TokenURI)
		log.Debugf("AnchorID: %s", hexutil.Encode(requestData.AnchorID.Bytes()))
		log.Debugf("NextAnchorID: %s", hexutil.Encode(requestData.NextAnchorID.Bytes()))
		log.Debugf("Props: %s", byteSlicetoString(requestData.Props))
		log.Debugf("Values: %s", byteSlicetoString(requestData.Values))
		log.Debugf("Salts: %s", byte32SlicetoString(requestData.Salts))
		log.Debugf("Proofs: %s", byteByte32SlicetoString(requestData.Proofs))

		isDone = <-done
		if !isDone {
			// some problem occurred in a child task
			errOut <- errors.New("mint nft failed for document %s and transaction %s", hexutil.Encode(req.DocumentID), utxID)
			return
		}

		// Check if tokenID exists in registry and owner is deposit address
		owner, err := s.OwnerOf(req.RegistryAddress, tokenID[:])
		if err != nil {
			errOut <- errors.New("error while checking new NFT owner %v", err)
			return
		}
		if owner.Hex() != req.DepositAddress.Hex() {
			errOut <- errors.New("Owner for tokenID %s should be %s, instead got %s", tokenID.String(), req.DepositAddress.Hex(), owner.Hex())
			return
		}

		log.Infof("Document %s minted successfully within transaction %s", hexutil.Encode(req.DocumentID), utxID)

		errOut <- nil
		return
	}
}

// OwnerOf returns the owner of the NFT token on ethereum chain
func (s *ethereumPaymentObligation) OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error) {
	contract, err := s.bindContract(registry, s.ethClient)
	if err != nil {
		return owner, errors.New("failed to bind the registry contract: %v", err)
	}

	opts, cancF := s.ethClient.GetGethCallOpts(false)
	defer cancF()

	return contract.OwnerOf(opts, utils.ByteSliceToBigInt(tokenID))
}

// MintRequest holds the data needed to mint and NFT from a Centrifuge document
type MintRequest struct {

	// To is the address of the recipient of the minted token
	To common.Address

	// TokenID is the ID for the minted token
	TokenID *big.Int

	// TokenURI is the metadata uri
	TokenURI string

	// AnchorID is the ID of the document as identified by the set up anchorRepository.
	AnchorID *big.Int

	// NextAnchorID is the next ID of the document, when updated
	NextAnchorID *big.Int

	// Props contains the compact props for readRole and tokenRole
	Props [][]byte

	// Values are the values of the leafs that is being proved Will be converted to string and concatenated for proof verification as outlined in precise-proofs library.
	Values [][]byte

	// salts are the salts for the field that is being proved Will be concatenated for proof verification as outlined in precise-proofs library.
	Salts [][32]byte

	// Proofs are the documents proofs that are needed
	Proofs [][][32]byte
}

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(tokenID TokenID, to common.Address, anchorID anchors.AnchorID, nextAnchorID anchors.AnchorID, proofs []*proofspb.Proof) (MintRequest, error) {
	proofData, err := convertToProofData(proofs)
	if err != nil {
		return MintRequest{}, err
	}

	return MintRequest{
		To:           to,
		TokenID:      tokenID.BigInt(),
		TokenURI:     tokenID.URI(),
		AnchorID:     anchorID.BigInt(),
		NextAnchorID: nextAnchorID.BigInt(),
		Props:        proofData.Props,
		Values:       proofData.Values,
		Salts:        proofData.Salts,
		Proofs:       proofData.Proofs}, nil
}

type proofData struct {
	Props  [][]byte
	Values [][]byte
	Salts  [][32]byte
	Proofs [][][32]byte
}

func convertToProofData(proofspb []*proofspb.Proof) (*proofData, error) {
	var props = make([][]byte, len(proofspb))
	var values = make([][]byte, len(proofspb))
	var salts = make([][32]byte, len(proofspb))
	var proofs = make([][][32]byte, len(proofspb))

	for i, p := range proofspb {
		salt32, err := utils.SliceToByte32(p.Salt)
		if err != nil {
			return nil, err
		}
		property, err := utils.ConvertProofForEthereum(p.SortedHashes)
		if err != nil {
			return nil, err
		}
		props[i] = p.GetCompactName()
		values[i] = p.Value
		// Scenario where it is a hashed field we copy the Hash value into the property value
		if len(p.Value) == 0 && len(p.Salt) == 0 {
			values[i] = p.Hash
		}
		salts[i] = salt32
		proofs[i] = property
	}

	return &proofData{Props: props, Values: values, Salts: salts, Proofs: proofs}, nil
}

func bindContract(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error) {
	return NewEthereumPaymentObligationContract(address, client.GetEthClient())
}

// Following are utility methods for nft parameter debugging purposes (Don't remove)

func byteSlicetoString(s [][]byte) string {
	str := "["

	for i := 0; i < len(s); i++ {
		str += "\"" + hexutil.Encode(s[i]) + "\",\n"
	}
	str += "]"
	return str
}

func byte32SlicetoString(s [][32]byte) string {
	str := "["

	for i := 0; i < len(s); i++ {
		str += "\"" + hexutil.Encode(s[i][:]) + "\",\n"
	}
	str += "]"
	return str
}

func byteByte32SlicetoString(s [][][32]byte) string {
	str := "["

	for i := 0; i < len(s); i++ {
		str += "\"" + byte32SlicetoString(s[i]) + "\",\n"
	}
	str += "]"
	return str
}
