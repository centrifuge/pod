package nft

import (
	"context"
	"math/big"
	"time"

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
	anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
	if err != nil {
		return mreq, err
	}

	nextAnchorID, err := anchors.ToAnchorID(model.NextVersion())
	if err != nil {
		return mreq, err
	}

	dr, err := model.CalculateDocumentRoot()
	if err != nil {
		return mreq, err
	}

	rootHash, err := anchors.ToDocumentRoot(dr)
	if err != nil {
		return mreq, err
	}

	requestData, err := NewMintRequest(model, tokenID, req.DepositAddress, anchorID, nextAnchorID, docProofs.FieldProofs, rootHash)
	if err != nil {
		return mreq, err
	}

	return requestData, nil

}

// MintNFT mints an NFT
func (s *ethereumPaymentObligation) MintNFT(ctx context.Context, req MintNFTRequest) (*MintNFTResponse, chan bool, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	cidBytes, err := tc.GetIdentityID()
	if err != nil {
		return nil, nil, err
	}

	cid := identity.NewDIDFromBytes(cidBytes)

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

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), cid, transactions.NilTxID(), "Minting NFT",
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
		tc, err := contextutil.Account(ctx)
		if err != nil {
			errOut <- err
			return
		}

		err = model.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:])
		if err != nil {
			errOut <- err
			return
		}

		_, _, done, err := s.docSrv.Update(contextutil.WithTX(ctx, txID), model)
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

		requestData, err := s.prepareMintRequest(ctx, tokenID, accountID, req)
		if err != nil {
			errOut <- errors.New("failed to prepare mint request: %v", err)
			return
		}

		opts, err := s.ethClient.GetTxOpts(tc.GetEthereumDefaultAccountName())
		if err != nil {
			errOut <- err
			return
		}

		contract, err := s.bindContract(req.RegistryAddress, s.ethClient)
		if err != nil {
			errOut <- err
			return
		}

		// to common.Address, tokenId *big.Int, tokenURI string, anchorId *big.Int, nextAnchorId *big.Int, properties [][]byte, values [][]byte, salts [][32]byte, proofs [][][32]byte
		ethTX, err := s.ethClient.SubmitTransactionWithRetries(contract.Mint, opts, requestData.To, requestData.TokenID,
			requestData.TokenURI, requestData.AnchorID, requestData.NextAnchorId, requestData.Props, requestData.Values,
			requestData.Salts, requestData.Proofs)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Sent off ethTX to mint [tokenID: %s, anchor: %x, nextAnchor: %s, registry: %s] to payment obligation contract. Ethereum transaction hash [%s] and Nonce [%d] and Check [%v]",
			requestData.TokenID, requestData.AnchorID, hexutil.Encode(requestData.NextAnchorId.Bytes()), requestData.To.String(), ethTX.Hash().String(), ethTX.Nonce(), ethTX.CheckNonce())
		log.Infof("Transfer pending: %s\n", ethTX.Hash().String())

		log.Debugf("To: %s", requestData.To.String())
		log.Debugf("TokenID: %s", hexutil.Encode(requestData.TokenID.Bytes()))
		log.Debugf("TokenURI: %s", requestData.TokenURI)
		log.Debugf("AnchorID: %s", hexutil.Encode(requestData.AnchorID.Bytes()))
		log.Debugf("NextAnchorID: %s", hexutil.Encode(requestData.NextAnchorId.Bytes()))
		log.Debugf("Props: %s", byteSlicetoString(requestData.Props))
		log.Debugf("Values: %s", byteSlicetoString(requestData.Values))
		log.Debugf("Salts: %s", byte32SlicetoString(requestData.Salts))
		log.Debugf("Proofs: %s", byteByte32SlicetoString(requestData.Proofs))

		res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), s.queue)
		if err != nil {
			errOut <- err
			return
		}

		_, err = res.Get(txMan.GetDefaultTaskTimeout())
		if err != nil {
			errOut <- err
			return
		}
		errOut <- nil
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

	// NextAnchorId is the next ID of the document, when updated
	NextAnchorId *big.Int

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
func NewMintRequest(model documents.Model, tokenID TokenID, to common.Address, anchorID anchors.AnchorID, nextAnchorID anchors.AnchorID, proofs []*proofspb.Proof, rootHash [32]byte) (MintRequest, error) {
	proofData, err := createProofData(model, proofs)
	if err != nil {
		return MintRequest{}, err
	}

	return MintRequest{
		To:           to,
		TokenID:      tokenID.BigInt(),
		TokenURI:     tokenID.URI(),
		AnchorID:     anchorID.BigInt(),
		NextAnchorId: nextAnchorID.BigInt(),
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

func createProofData(model documents.Model, proofspb []*proofspb.Proof) (*proofData, error) {
	// TODO cleanup hard coded indexes using the model props
	readRoleIndex := 5
	tokenRoleIndex := 7
	var props = make([][]byte, 2)  // props are only required for readRole.property, tokenRole.property
	var values = make([][]byte, 4) // values are only required for readRole.Value
	var salts = make([][32]byte, len(proofspb))
	var proofs = make([][][32]byte, len(proofspb))

	// TODO remove later
	//proof, _ := documents.ConvertDocProofToClientFormat(&documents.DocumentProof{FieldProofs: proofspb})
	//log.Info(json.MarshalIndent(proof, "", "  "))
	for i, p := range proofspb {
		if i == readRoleIndex {
			props[0] = p.GetCompactName()
		}
		if i == tokenRoleIndex {
			props[1] = p.GetCompactName()
		}

		if i < 3 {
			values[i] = p.Value
		}

		if i == readRoleIndex {
			values[3] = p.Value
		}

		salt32, err := utils.SliceToByte32(p.Salt)
		if err != nil {
			return nil, err
		}

		salts[i] = salt32
		property, err := convertProofProperty(p.SortedHashes)
		if err != nil {
			return nil, err
		}
		proofs[i] = property
	}

	return &proofData{Props: props, Values: values, Salts: salts, Proofs: proofs}, nil
}

func convertProofProperty(sortedHashes [][]byte) ([][32]byte, error) {
	var property [][32]byte
	for _, hash := range sortedHashes {
		hash32, err := utils.SliceToByte32(hash)
		if err != nil {
			return nil, err
		}
		property = append(property, hash32)
	}

	return property, nil
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
