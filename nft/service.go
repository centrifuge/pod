package nft

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/precise-proofs/proofs"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
	"golang.org/x/crypto/sha3"
)

var log = logging.Logger("nft")

const (
	// ErrNFTMinted error for NFT already minted for registry
	ErrNFTMinted = errors.Error("NFT already minted")

	// GenericMintMethodABI constant interface to interact with mint methods
	GenericMintMethodABI = `[{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tkn","type":"uint256"},{"internalType":"bytes32","name":"dataRoot","type":"bytes32"},{"internalType":"bytes[]","name":"properties","type":"bytes[]"},{"internalType":"bytes[]","name":"values","type":"bytes[]"},{"internalType":"bytes32[]","name":"salts","type":"bytes32[]"}],"name":"mint","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

	// AssetStoredEventSignature used for finding events
	AssetStoredEventSignature = "AssetStored(bytes32)"

	// ABI is string abi with required methods to call the NFT registry contract
	ABI = `[{"constant":true,"inputs":[{"name":"tokenId","type":"uint256"}],"name":"currentIndexOfToken","outputs":[{"name":"index","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
)

// nftABI is the default abi for caller functions on NFT registry
var nftABI abi.ABI

func init() {
	var err error
	nftABI, err = abi.JSON(strings.NewReader(ABI))
	if err != nil {
		log.Fatalf("failed to decode NFT ABI: %v", err)
	}
	gob.Register(TokenID{})
	gob.Register(MintNFTRequest{})
	gob.Register(new(big.Int))
	gob.Register(MintRequest{})
	gob.Register(gocelery.JobID{})
	gob.Register(common.Address{})
}

// Config is the config interface for nft package
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
	GetLowEntropyNFTTokenEnabled() bool
}

// service handles all interactions related to minting of NFTs for unpaid invoices on Ethereum
type service struct {
	cfg                Config
	identityService    identity.Service
	ethClient          ethereum.Client
	queue              queue.TaskQueuer
	docSrv             documents.Service
	bindCallerContract func(address common.Address, abi abi.ABI, client ethereum.Client) *bind.BoundContract
	jobsManager        jobs.Manager
	dispatcher         jobsv2.Dispatcher
	api                API
	blockHeightFunc    func() (height uint64, err error)
}

// newService creates InvoiceUnpaid given the parameters
func newService(
	cfg Config,
	identityService identity.Service,
	ethClient ethereum.Client,
	queue queue.TaskQueuer,
	docSrv documents.Service,
	bindCallerContract func(address common.Address, abi abi.ABI, client ethereum.Client) *bind.BoundContract,
	jobsMan jobs.Manager,
	dispatcher jobsv2.Dispatcher,
	api API,
	blockHeightFunc func() (uint64, error)) *service {
	return &service{
		cfg:                cfg,
		identityService:    identityService,
		ethClient:          ethClient,
		queue:              queue,
		docSrv:             docSrv,
		bindCallerContract: bindCallerContract,
		jobsManager:        jobsMan,
		dispatcher:         dispatcher,
		blockHeightFunc:    blockHeightFunc,
		api:                api,
	}
}

func prepareMintRequest(ctx context.Context, docSrv documents.Service, tokenID TokenID, cid identity.DID,
	req MintNFTRequest) (mreq MintRequest, err error) {
	docProofs, err := docSrv.CreateProofs(ctx, req.DocumentID, req.ProofFields)
	if err != nil {
		return mreq, err
	}

	model, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)
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

	docProofs.FieldProofs = append(docProofs.FieldProofs, pfs.FieldProofs...)
	signaturesRoot, err := model.CalculateSignaturesRoot()
	if err != nil {
		return mreq, err
	}
	signingRoot, err := model.CalculateSigningRoot()
	if err != nil {
		return mreq, err
	}

	anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
	if err != nil {
		return mreq, err
	}

	nextAnchorID, err := anchors.ToAnchorID(model.NextVersion())
	if err != nil {
		return mreq, err
	}

	docRoot, err := model.CalculateDocumentRoot()
	if err != nil {
		return mreq, err
	}

	optProofs, err := proofs.OptimizeProofs(docProofs.FieldProofs, docRoot, sha3.NewLegacyKeccak256())
	if err != nil {
		return mreq, err
	}

	// useful to log proof data to be passed to mint method
	log.Debugf("\nDocumentRoot %x\nSignaturesRoot %x\nSigningRoot %x\nDocumentID %x\nCurrentVersion %x\n",
		docRoot, signaturesRoot, signingRoot, model.ID(), model.CurrentVersion())
	log.Debug(json.MarshalIndent(documents.ConvertProofs(optProofs), "", "  "))

	requestData, err := NewMintRequest(tokenID, req.DepositAddress, anchorID, nextAnchorID, docProofs.LeftDataRooot, docProofs.RightDataRoot, signingRoot, signaturesRoot, optProofs)
	if err != nil {
		return mreq, err
	}

	return requestData, nil
}

// MintNFT mints an NFT
func (s *service) MintNFT(ctx context.Context, req MintNFTRequest) (*TokenResponse, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	if !req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof {
		return nil, errors.New("enable grant_nft_access to generate Read Access Proof")
	}

	tokenID := NewTokenID()
	if s.cfg.GetLowEntropyNFTTokenEnabled() {
		log.Warnf("Security consideration: Using a reduced maximum of %s integer for NFT token ID generation. "+
			"Suggested course of action: disable by setting nft.lowentropy=false in config.yaml file", LowEntropyTokenIDMax)
		tokenID = NewLowEntropyTokenID()
	}

	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return nil, err
	}

	// check if the nft is successfully minted already
	if model.IsNFTMinted(s, req.RegistryAddress) {
		return nil, errors.NewTypedError(ErrNFTMinted, errors.New("registry %v", req.RegistryAddress.String()))
	}

	didBytes := tc.GetIdentityID()

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, err
	}

	jobID, err := initiateNFTMint(s.dispatcher, did, tokenID, req)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		JobID:   jobID.Hex(),
		TokenID: tokenID.String(),
	}, nil
}

// TransferFrom transfers an NFT to another address
func (s *service) TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID TokenID) (*TokenResponse, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	didBytes := tc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, err
	}

	owner, err := s.OwnerOf(registry, tokenID[:])
	if err != nil {
		return nil, fmt.Errorf("failed to get current owner: %w", err)
	}

	if !bytes.Equal(owner.Bytes(), did.ToAddress().Bytes()) {
		return nil, fmt.Errorf("%s is not the owner of NFT[%s]", did, tokenID)
	}

	jobID, err := initiateTransferNFTJob(s.dispatcher, did, to, registry, tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch transfer nft job: %w", err)
	}

	return &TokenResponse{
		JobID:   jobID.Hex(),
		TokenID: tokenID.String(),
	}, nil
}

// OwnerOf returns the owner of the NFT token on ethereum chain
func (s *service) OwnerOf(registry common.Address, tokenID []byte) (common.Address, error) {
	return ownerOf(s.ethClient, registry, tokenID)
}

func ownerOf(ethClient ethereum.Client, registry common.Address, tokenID []byte) (common.Address, error) {
	var owner common.Address
	var err error

	c := ethereum.BindContract(registry, nftABI, ethClient)
	opts, cancF := ethClient.GetGethCallOpts(false)
	defer cancF()

	err = c.Call(opts, &owner, "ownerOf", utils.ByteSliceToBigInt(tokenID))
	if err != nil {
		log.Warnf("Error getting NFT owner for token [%x]: %v", tokenID, err)
	}

	return owner, err
}

// CurrentIndexOfToken returns the current index of the token in the given registry
func (s *service) CurrentIndexOfToken(registry common.Address, tokenID []byte) (*big.Int, error) {
	c := s.bindCallerContract(registry, nftABI, s.ethClient)
	opts, cancF := s.ethClient.GetGethCallOpts(false)
	defer cancF()

	res := new(big.Int)
	return res, c.Call(opts, res, "currentIndexOfToken", utils.ByteSliceToBigInt(tokenID))
}

// MintRequest holds the data needed to mint and NFT from a Centrifuge document
type MintRequest struct {

	// To is the address of the recipient of the minted token
	To common.Address

	// TokenID is the ID for the minted token
	TokenID *big.Int

	// AnchorID is the ID of the document as identified by the set up anchorRepository.
	AnchorID anchors.AnchorID

	// NextAnchorID is the next ID of the document, when updated
	NextAnchorID *big.Int

	// LeftDataRoot of the document
	LeftDataRoot [32]byte

	// RightDataRoot of the document
	RightDataRoot [32]byte

	// SigningRoot of the document
	SigningRoot [32]byte

	// SignaturesRoot of the document
	SignaturesRoot [32]byte

	// Props contains the compact props for readRole and tokenRole
	Props [][]byte

	// Values are the values of the leafs that is being proved Will be converted to string and concatenated for proof verification as outlined in precise-proofs library.
	Values [][]byte

	// salts are the salts for the field that is being proved Will be concatenated for proof verification as outlined in precise-proofs library.
	Salts [][32]byte

	// Proofs are the documents proofs that are needed
	Proofs [][][32]byte

	// bundled hash is the keccak hash of to + (props+values+salts)
	BundledHash [32]byte

	// static proofs holds data root, sibling root and signature root
	StaticProofs [3][32]byte
}

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(
	tokenID TokenID,
	to common.Address,
	anchorID anchors.AnchorID,
	nextAnchorID anchors.AnchorID,
	leftDataRoot, rightDataRoot, signingRoot, signaturesRoot []byte,
	proofs []*proofspb.Proof) (MintRequest, error) {
	proofData, err := convertToProofData(proofs)
	if err != nil {
		return MintRequest{}, err
	}
	ldr := utils.MustSliceToByte32(leftDataRoot)
	rdr := utils.MustSliceToByte32(rightDataRoot)
	snr := utils.MustSliceToByte32(signingRoot)
	sgr := utils.MustSliceToByte32(signaturesRoot)
	bh := getBundledHash(to, proofData.Props, proofData.Values, proofData.Salts)
	return MintRequest{
		To:             to,
		TokenID:        tokenID.BigInt(),
		AnchorID:       anchorID,
		NextAnchorID:   nextAnchorID.BigInt(),
		LeftDataRoot:   ldr,
		RightDataRoot:  rdr,
		SigningRoot:    snr,
		SignaturesRoot: sgr,
		Props:          proofData.Props,
		Values:         proofData.Values,
		Salts:          proofData.Salts,
		Proofs:         proofData.Proofs,
		BundledHash:    bh}, nil
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
	var pfs = make([][][32]byte, len(proofspb))

	for i, p := range proofspb {
		salt32, err := utils.SliceToByte32(p.Salt)
		if err != nil {
			return nil, err
		}
		property, err := utils.ConvertProofForEthereum(p.SortedHashes)
		if err != nil {
			return nil, err
		}

		props[i] = proofs.AsBytes(p.Property)
		values[i] = p.Value
		// Scenario where it is a hashed field we copy the Hash value into the property value
		if len(p.Value) == 0 && len(p.Salt) == 0 {
			values[i] = p.Hash
		}
		salts[i] = salt32
		pfs[i] = property
	}

	return &proofData{Props: props, Values: values, Salts: salts, Proofs: pfs}, nil
}

// getBundledHash returns the sha3 of the concat of to + (props+values+salts)
func getBundledHash(to common.Address, props, values [][]byte, salts [][32]byte) [32]byte {
	res := to.Bytes()
	for i := 0; i < len(props); i++ {
		// keccak256(prop[i]+values[i]+salts[i])
		h := getLeafHash(props[i], values[i], salts[i])

		// append h to res
		res = append(res, h...)
	}

	// return keccak256(res)
	return utils.MustSliceToByte32(crypto.Keccak256(res))
}

func getLeafHash(prop, value []byte, salt [32]byte) []byte {
	// append prop+value+salt
	h := append(prop, value...)
	h = append(h, salt[:]...)
	return crypto.Keccak256(h)
}
