package nft

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"time"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("nft")


// Config is an interface to configurations required by nft package
type Config interface {
	GetIdentityID() ([]byte, error)
	GetEthereumDefaultAccountName() string
	GetContractAddress(address string) common.Address
	GetEthereumContextWaitTimeout() time.Duration
}

// ethereumPaymentObligationContract is an abstraction over the contract code to help in mocking it out
type ethereumPaymentObligationContract interface {

	// Mint method abstracts Mint method on the contract
	Mint(opts *bind.TransactOpts, to common.Address, tokenID *big.Int, tokenURI string, anchorID *big.Int, merkleRoot [32]byte, values []string, salts [][32]byte, proofs [][][32]byte) (*types.Transaction, error)
}

// ethereumPaymentObligation handles all interactions related to minting of NFTs for payment obligations on Ethereum
type ethereumPaymentObligation struct {
	registry          *documents.ServiceRegistry
	identityService   identity.Service
	ethClient         ethereum.Client
	config            Config
	queue             *queue.Server
	setupMintListener func(config Config, queue *queue.Server, tokenID *big.Int, registryAddress string) (confirmations chan *watchTokenMinted, err error)
	bindContract      func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error)
}

// newEthereumPaymentObligation creates ethereumPaymentObligation given the parameters
func newEthereumPaymentObligation(registry *documents.ServiceRegistry, identityService identity.Service, ethClient ethereum.Client, config Config,
	queue *queue.Server,
	setupMintListener func(config Config, queue *queue.Server, tokenID *big.Int, registryAddress string) (confirmations chan *watchTokenMinted, err error), bindContract func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error)) *ethereumPaymentObligation {
	return &ethereumPaymentObligation{
		registry:          registry,
		identityService:   identityService,
		ethClient:         ethClient,
		config:            config,
		setupMintListener: setupMintListener,
		bindContract:      bindContract,
		queue:             queue,
	}
}

func (s *ethereumPaymentObligation) prepareMintRequest(documentID []byte, depositAddress string, proofFields []string) (*MintRequest, error) {
	docService, err := s.registry.FindService(documentID)
	if err != nil {
		return nil, err
	}

	model, err := docService.GetCurrentVersion(documentID)
	if err != nil {
		return nil, err
	}

	corDoc, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	proofs, err := docService.CreateProofs(documentID, proofFields)
	if err != nil {
		return nil, err
	}

	toAddress := common.HexToAddress(depositAddress)

	anchorID, err := anchors.ToAnchorID(corDoc.CurrentVersion)
	if err != nil {
		return nil, nil
	}

	rootHash, err := anchors.ToDocumentRoot(corDoc.DocumentRoot)
	if err != nil {
		return nil, nil
	}


	requestData, err := NewMintRequest(toAddress, anchorID, proofs.FieldProofs, rootHash)
	if err != nil {
		return nil, err
	}

	return requestData, nil

}

// MintNFT mints an NFT
func (s *ethereumPaymentObligation) MintNFT(documentID []byte, registryAddress, depositAddress string, proofFields []string) (<-chan *watchTokenMinted, error) {

	requestData, err := s.prepareMintRequest(documentID, depositAddress, proofFields)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare mint request: %v", err)
	}

	opts, err := s.ethClient.GetTxOpts(s.config.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, err
	}

	var contract *EthereumPaymentObligationContract
	if registryAddress == "" {
		defaultRegistry := s.config.GetContractAddress("paymentObligation")
		contract, err = s.bindContract(defaultRegistry, s.ethClient)
		registryAddress = defaultRegistry.String()
	} else {
		contract, err = s.bindContract(common.HexToAddress(registryAddress), s.ethClient)
	}
	if err != nil {
		return nil, err
	}

	watch, err := s.setupMintListener(s.config, s.queue, requestData.TokenID, registryAddress)
	if err != nil {
		return nil, err
	}

	err = s.sendMintTransaction(contract, opts, requestData)
	if err != nil {
		return nil, err
	}

	return watch, nil
}

// setUpMintEventListener sets up the listened for the "PaymentObligationMinted" event to notify the upstream code
// about successful minting of an NFt
func setupMintListener(config Config, qs *queue.Server, tokenID *big.Int, registryAddress string) (confirmations chan *watchTokenMinted, err error) {
	confirmations = make(chan *watchTokenMinted)
	conn := ethereum.GetClient()

	h, err := conn.GetEthClient().HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	asyncRes, err := qs.EnqueueJob(mintingConfirmationTaskName, map[string]interface{}{
		tokenIDParam:           hex.EncodeToString(tokenID.Bytes()),
		queue.BlockHeightParam: h.Number.Uint64(),
		registryAddressParam:   registryAddress,
	})
	if err != nil {
		return nil, err
	}

	go waitAndRouteNFTApprovedEvent(config.GetEthereumContextWaitTimeout(), asyncRes, tokenID, confirmations)
	return confirmations, nil
}

// waitAndRouteNFTApprovedEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteNFTApprovedEvent(timeout time.Duration, asyncRes queue.TaskResult, tokenID *big.Int, confirmations chan<- *watchTokenMinted) {
	_, err := asyncRes.Get(timeout)
	confirmations <- &watchTokenMinted{tokenID, err}
}

// sendMintTransaction sends the actual transaction to mint the NFT
func (s *ethereumPaymentObligation) sendMintTransaction(contract ethereumPaymentObligationContract, opts *bind.TransactOpts, requestData *MintRequest) error {
	tx, err := s.ethClient.SubmitTransactionWithRetries(contract.Mint, opts, requestData.To, requestData.TokenID, requestData.TokenURI, requestData.AnchorID,
		requestData.MerkleRoot, requestData.Values, requestData.Salts, requestData.Proofs)
	if err != nil {
		return err
	}
	log.Infof("Sent off tx to mint [tokenID: %x, anchor: %x, registry: %x] to payment obligation contract. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]",
		requestData.TokenID, requestData.AnchorID, requestData.To, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
	return nil
}

func (s *ethereumPaymentObligation) getIdentityAddress() (common.Address, error) {
	centIDBytes, err := s.config.GetIdentityID()
	if err != nil {
		return common.Address{}, err
	}

	centID, err := identity.ToCentID(centIDBytes)
	if err != nil {
		return common.Address{}, err
	}

	address, err := s.identityService.GetIdentityAddress(centID)
	if err != nil {
		return common.Address{}, err
	}
	return address, nil
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
func NewMintRequest(to common.Address, anchorID anchors.AnchorID, proofs []*proofspb.Proof, rootHash [32]byte) (*MintRequest, error) {
	tokenID := utils.ByteSliceToBigInt(utils.RandomSlice(256))
	tokenURI := "http:=//www.centrifuge.io/DUMMY_URI_SERVICE"
	proofData, err := createProofData(proofs)
	if err != nil {
		return nil, err
	}

	return &MintRequest{
		To:                to,
		TokenID:           tokenID,
		TokenURI:          tokenURI,
		AnchorID:          anchorID.BigInt(),
		MerkleRoot:        rootHash,
		Values:            proofData.Values,
		Salts:             proofData.Salts,
		Proofs:            proofData.Proofs}, nil
}

type proofData struct {
	Values []string
	Salts  [][32]byte
	Proofs [][][32]byte
}

func createProofData(proofspb []*proofspb.Proof) (*proofData, error) {
	var values = make([]string, len(proofspb))
	var salts = make([][32]byte, len(proofspb))
	var proofs =  make([][][32]byte, len(proofspb))

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
