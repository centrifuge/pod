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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

var log = logging.Logger("nft")

// ethereumPaymentObligationContract is an abstraction over the contract code to help in mocking it out
type ethereumPaymentObligationContract interface {

	// Mint method abstracts Mint method on the contract
	Mint(opts *bind.TransactOpts, to common.Address, tokenID *big.Int, tokenURI string, anchorID *big.Int, merkleRoot [32]byte, values []string, salts [][32]byte, proofs [][][32]byte) (*types.Transaction, error)

	// OwnerOf to retrieve owner of the tokenID
	OwnerOf(opts *bind.CallOpts, tokenID *big.Int) (common.Address, error)
}

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
	txService       transactions.Manager
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
	txService transactions.Manager,
	blockHeightFunc func() (uint64, error)) *ethereumPaymentObligation {
	return &ethereumPaymentObligation{
		cfg:             cfg,
		identityService: identityService,
		ethClient:       ethClient,
		bindContract:    bindContract,
		queue:           queue,
		docSrv:          docSrv,
		txService:       txService,
		blockHeightFunc: blockHeightFunc,
	}
}

func (s *ethereumPaymentObligation) prepareMintRequest(ctx context.Context, documentID []byte, depositAddress string, proofFields []string) (MintRequest, error) {
	model, err := s.docSrv.GetCurrentVersion(ctx, documentID)
	if err != nil {
		return MintRequest{}, err
	}

	corDoc, err := model.PackCoreDocument()
	if err != nil {
		return MintRequest{}, err
	}

	proofs, err := s.docSrv.CreateProofs(ctx, documentID, proofFields)
	if err != nil {
		return MintRequest{}, err
	}

	toAddress := common.HexToAddress(depositAddress)

	anchorID, err := anchors.ToAnchorID(corDoc.CurrentVersion)
	if err != nil {
		return MintRequest{}, nil
	}

	rootHash, err := anchors.ToDocumentRoot(corDoc.DocumentRoot)
	if err != nil {
		return MintRequest{}, nil
	}

	requestData, err := NewMintRequest(toAddress, anchorID, proofs.FieldProofs, rootHash)
	if err != nil {
		return MintRequest{}, err
	}

	return requestData, nil

}

// MintNFT mints an NFT
func (s *ethereumPaymentObligation) MintNFT(ctx context.Context, documentID []byte, registryAddress, depositAddress string, proofFields []string) (*MintNFTResponse, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	requestData, err := s.prepareMintRequest(ctx, documentID, depositAddress, proofFields)
	if err != nil {
		return nil, errors.New("failed to prepare mint request: %v", err)
	}

	opts, err := s.ethClient.GetTxOpts(tc.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, err
	}

	registry := common.HexToAddress(registryAddress)
	contract, err := s.bindContract(registry, s.ethClient)
	if err != nil {
		return nil, err
	}

	cidBytes, err := tc.GetIdentityID()
	if err != nil {
		return nil, err
	}

	cid, err := identity.ToCentID(cidBytes)
	if err != nil {
		return nil, err
	}

	txID, err := s.sendMintTransaction(cid, contract, opts, requestData, registry, documentID)
	if err != nil {
		return nil, errors.New("failed to send transaction: %v", err)
	}

	return &MintNFTResponse{
		TransactionID: txID.String(),
		TokenID:       requestData.TokenID.String(),
	}, nil
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

// sendMintTransaction sends the actual transaction to mint the NFT
func (s *ethereumPaymentObligation) sendMintTransaction(
	cid identity.CentID,
	contract ethereumPaymentObligationContract,
	opts *bind.TransactOpts,
	requestData MintRequest,
	registry common.Address,
	documentID []byte) (uuid.UUID, error) {

	r := requestData.copy()

	// Run within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	txID, _, err := s.txService.ExecuteWithinTX(context.Background(), cid, uuid.Nil, "Minting NFT", func(accountID identity.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		ethTX, err := s.ethClient.SubmitTransactionWithRetries(contract.Mint, opts, r.To, r.TokenID, r.TokenURI, r.AnchorID,
			r.MerkleRoot, r.Values, r.Salts, r.Proofs)
		if err != nil {
			errOut <- err
		}

		log.Infof("Sent off ethTX to mint [tokenID: %x, anchor: %x, registry: %x] to payment obligation contract. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]",
			requestData.TokenID, requestData.AnchorID, requestData.To, ethTX.Hash(), ethTX.Nonce(), ethTX.CheckNonce())
		log.Infof("Transfer pending: 0x%x\n", ethTX.Hash())

		res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), true, s.queue)
		if err != nil {
			errOut <- err
		}

		_, err = res.Get(s.cfg.GetEthereumContextWaitTimeout())
		if err != nil {
			errOut <- err
		}

		res, err = documents.InitNFTCreatedTask(
			s.queue, txID, accountID, documentID, registry, requestData.TokenID.Bytes())
		if err != nil {
			errOut <- err
		}

		_, err = res.Get(s.cfg.GetEthereumContextWaitTimeout())
		if err != nil {
			errOut <- err
		}

		errOut <- nil
	})

	return txID, err
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
	Values []string

	// salts are the salts for the field that is being proved Will be concatenated for proof verification as outlined in precise-proofs library.
	Salts [][32]byte

	// Proofs are the documents proofs that are needed
	Proofs [][][32]byte
}

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(to common.Address, anchorID anchors.AnchorID, proofs []*proofspb.Proof, rootHash [32]byte) (MintRequest, error) {

	// tokenID is uint256 in Solidity (256 bits | max value is 2^256-1)
	// tokenID should be random 32 bytes (32 byte = 256 bits)
	tokenID := utils.ByteSliceToBigInt(utils.RandomSlice(32))
	tokenURI := "http:=//www.centrifuge.io/DUMMY_URI_SERVICE"
	proofData, err := createProofData(proofs)
	if err != nil {
		return MintRequest{}, err
	}

	return MintRequest{
		To:         to,
		TokenID:    tokenID,
		TokenURI:   tokenURI,
		AnchorID:   anchorID.BigInt(),
		MerkleRoot: rootHash,
		Values:     proofData.Values,
		Salts:      proofData.Salts,
		Proofs:     proofData.Proofs}, nil
}

func (m MintRequest) copy() MintRequest {
	return MintRequest{
		m.To,
		new(big.Int).Set(m.TokenID),
		m.TokenURI,
		new(big.Int).Set(m.AnchorID),
		m.MerkleRoot,
		m.Values,
		m.Salts,
		m.Proofs,
	}
}

type proofData struct {
	Values []string
	Salts  [][32]byte
	Proofs [][][32]byte
}

func createProofData(proofspb []*proofspb.Proof) (*proofData, error) {
	var values = make([]string, len(proofspb))
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
