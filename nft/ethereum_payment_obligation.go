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
	"github.com/satori/go.uuid"
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
	identityService identity.Service
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
	identityService identity.Service,
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

func (s *ethereumPaymentObligation) prepareMintRequest(ctx context.Context, tokenID TokenID, cid identity.CentID, req MintNFTRequest) (mreq MintRequest, err error) {
	docProofs, err := s.docSrv.CreateProofs(ctx, req.DocumentID, req.ProofFields)
	if err != nil {
		return mreq, err
	}

	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return mreq, err
	}

	coreDoc, err := model.PackCoreDocument()
	if err != nil {
		return mreq, err
	}

	dataRoot, err := model.CalculateDataRoot()
	if err != nil {
		return mreq, err
	}

	pfs, err := coreDoc.GetNFTProofs(
		dataRoot,
		cid,
		req.RegistryAddress,
		tokenID[:],
		req.SubmitRoleProof,
		req.SubmitTokenProof,
		req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof)
	if err != nil {
		return mreq, err
	}

	docProofs.FieldProofs = append(docProofs.FieldProofs, pfs...)
	anchorID, err := anchors.ToAnchorID(coreDoc.Document.CurrentVersion)
	if err != nil {
		return mreq, err
	}

	rootHash, err := anchors.ToDocumentRoot(coreDoc.Document.DocumentRoot)
	if err != nil {
		return mreq, err
	}

	requestData, err := NewMintRequest(tokenID, req.DepositAddress, anchorID, pfs, rootHash)
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

	cid, err := identity.ToCentID(cidBytes)
	if err != nil {
		return nil, nil, err
	}

	if !req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof {
		return nil, nil, errors.New("enabled grant nft access to generate NFT read access proof")
	}

	tokenID := NewTokenID()
	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return nil, nil, err
	}

	dm, err := model.PackCoreDocument()
	if err != nil {
		return nil, nil, err
	}

	if req.SubmitRoleProof != nil && !dm.IsAccountInRole(req.SubmitRoleProof, cid) {
		return nil, nil, documents.ErrNFTRoleMissing
	}

	// check if the nft is successfully minted already
	if dm.IsNFTMinted(s, req.RegistryAddress) {
		return nil, nil, errors.NewTypedError(ErrNFTMinted, errors.New("registry %v", req.RegistryAddress.String()))
	}

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), cid, uuid.Nil, "Minting NFT",
		s.minter(ctx, tokenID, model, req))
	if err != nil {
		return nil, nil, err
	}

	return &MintNFTResponse{
		TransactionID: txID.String(),
		TokenID:       tokenID.String(),
	}, done, nil
}

func (s *ethereumPaymentObligation) minter(ctx context.Context, tokenID TokenID, model documents.Model, req MintNFTRequest) func(accountID identity.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID identity.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		tc, err := contextutil.Account(ctx)
		if err != nil {
			errOut <- err
			return
		}

		dm, err := model.PackCoreDocument()
		if err != nil {
			errOut <- err
			return
		}

		ndm, err := dm.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:])
		if err != nil {
			errOut <- err
			return
		}

		model, err = s.docSrv.DeriveFromCoreDocumentModel(ndm)
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
			// some problem occured in a child task
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

		ethTX, err := s.ethClient.SubmitTransactionWithRetries(contract.Mint, opts, requestData.To, requestData.TokenID, requestData.TokenURI, requestData.AnchorID,
			requestData.MerkleRoot, requestData.Values, requestData.Salts, requestData.Proofs)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Sent off ethTX to mint [tokenID: %s, anchor: %x, registry: %s] to payment obligation contract. Ethereum transaction hash [%s] and Nonce [%d] and Check [%v]",
			requestData.TokenID, requestData.AnchorID, requestData.To.String(), ethTX.Hash().String(), ethTX.Nonce(), ethTX.CheckNonce())
		log.Infof("Transfer pending: %s\n", ethTX.Hash().String())

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

	// MerkleRoot is the root hash of the merkle proof/doc
	MerkleRoot [32]byte

	// Values are the values of the leafs that is being proved Will be converted to string and concatenated for proof verification as outlined in precise-proofs library.
	Values [][]byte

	// salts are the salts for the field that is being proved Will be concatenated for proof verification as outlined in precise-proofs library.
	Salts [][32]byte

	// Proofs are the documents proofs that are needed
	Proofs [][][32]byte
}

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(tokenID TokenID, to common.Address, anchorID anchors.AnchorID, proofs []*proofspb.Proof, rootHash [32]byte) (MintRequest, error) {
	proofData, err := createProofData(proofs)
	if err != nil {
		return MintRequest{}, err
	}

	return MintRequest{
		To:         to,
		TokenID:    tokenID.BigInt(),
		TokenURI:   tokenID.URI(),
		AnchorID:   anchorID.BigInt(),
		MerkleRoot: rootHash,
		Values:     proofData.Values,
		Salts:      proofData.Salts,
		Proofs:     proofData.Proofs}, nil
}

type proofData struct {
	Values [][]byte
	Salts  [][32]byte
	Proofs [][][32]byte
}

func createProofData(proofspb []*proofspb.Proof) (*proofData, error) {
	var values = make([][]byte, len(proofspb))
	var salts = make([][32]byte, len(proofspb))
	var proofs = make([][][32]byte, len(proofspb))

	for i, p := range proofspb {
		values[i] = p.Value
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

	return &proofData{Values: values, Salts: salts, Proofs: proofs}, nil
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
