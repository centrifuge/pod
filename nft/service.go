package nft

import (
	"context"
	"encoding/json"
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
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("nft")

const (
	// ErrNFTMinted error for NFT already minted for registry
	ErrNFTMinted = errors.Error("NFT already minted")

	// GenericMintMethodABI constant interface to interact with mint methods
	GenericMintMethodABI = `[{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"bytes[]","name":"properties","type":"bytes[]"},{"internalType":"bytes[]","name":"values","type":"bytes[]"},{"internalType":"bytes32[]","name":"salts","type":"bytes32[]"}],"name":"mint","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

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
		blockHeightFunc:    blockHeightFunc,
		api:                api,
	}
}

func (s *service) prepareMintRequest(ctx context.Context, tokenID TokenID, cid identity.DID, req MintNFTRequest) (mreq MintRequest, err error) {
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

	optProofs, err := proofs.OptimizeProofs(docProofs.FieldProofs, docRoot, sha3.NewKeccak256())
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
func (s *service) MintNFT(ctx context.Context, req MintNFTRequest) (*TokenResponse, chan error, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	if !req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof {
		return nil, nil, errors.New("enable grant_nft_access to generate Read Access Proof")
	}

	tokenID := NewTokenID()
	if s.cfg.GetLowEntropyNFTTokenEnabled() {
		log.Warningf("Security consideration: Using a reduced maximum of %s integer for NFT token ID generation. "+
			"Suggested course of action: disable by setting nft.lowentropy=false in config.yaml file", LowEntropyTokenIDMax)
		tokenID = NewLowEntropyTokenID()
	}

	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return nil, nil, err
	}

	// check if the nft is successfully minted already
	if model.IsNFTMinted(s, req.RegistryAddress) {
		return nil, nil, errors.NewTypedError(ErrNFTMinted, errors.New("registry %v", req.RegistryAddress.String()))
	}

	didBytes := tc.GetIdentityID()

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, nil, err
	}

	jobID, done, err := s.jobsManager.ExecuteWithinJob(contextutil.Copy(ctx), did, jobs.NilJobID(), "Minting NFT",
		s.minterJob(ctx, tokenID, model, req))

	if err != nil {
		return nil, nil, err
	}

	return &TokenResponse{
		JobID:   jobID.String(),
		TokenID: tokenID.String(),
	}, done, nil
}

// TransferFrom transfers an NFT to another address
func (s *service) TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID TokenID) (*TokenResponse, chan error, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	didBytes := tc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, nil, err
	}

	jobID, done, err := s.jobsManager.ExecuteWithinJob(contextutil.Copy(ctx), did, jobs.NilJobID(), "Transfer From NFT",
		s.transferFromJob(ctx, registry, did.ToAddress(), to, tokenID))
	if err != nil {
		return nil, nil, err
	}

	return &TokenResponse{
		JobID:   jobID.String(),
		TokenID: tokenID.String(),
	}, done, nil
}

func (s *service) minterJob(ctx context.Context, tokenID TokenID, model documents.Model, req MintNFTRequest) func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
		err := model.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:])
		if err != nil {
			errOut <- err
			return
		}

		jobCtx := contextutil.WithJob(ctx, jobID)
		_, _, done, err := s.docSrv.Update(jobCtx, model)
		if err != nil {
			errOut <- err
			return
		}

		err = <-done
		if err != nil {
			// some problem occurred in a child task
			errOut <- errors.New("update document failed for document %s and job %s with error %s", hexutil.Encode(req.DocumentID), jobID, err.Error())
			return
		}

		requestData, err := s.prepareMintRequest(jobCtx, tokenID, accountID, req)
		if err != nil {
			errOut <- errors.New("failed to prepare mint request: %v", err)
			return
		}

		subProofs := toSubstrateProofs(requestData.Props, requestData.Values, requestData.Salts, requestData.Proofs)
		staticProofs := [3][32]byte{requestData.LeftDataRoot, requestData.RightDataRoot, requestData.SignaturesRoot}
		block, err := s.ethClient.GetEthClient().BlockByNumber(context.Background(), nil)
		if err != nil {
			errOut <- errors.New("failed to get latest block: %v", err)
			return
		}

		done, err = s.api.ValidateNFT(ctx, requestData.AnchorID, requestData.To, subProofs, staticProofs)
		if err != nil {
			errOut <- err
			return
		}

		if err := <-done; err != nil {
			errOut <- err
			return
		}
		log.Infof("Successfully validated Proofs on cent chain for anchorID: %s", requestData.AnchorID.String())

		if !utils.IsEmptyAddress(req.AssetManagerAddress) {
			// listen for event
			txHash, done, err := ethereum.CreateWaitForEventJob(
				ctx, txMan, s.queue, accountID, jobID,
				AssetStoredEventSignature, block.Number(), req.AssetManagerAddress, requestData.BundledHash)
			if err != nil {
				errOut <- err
				return
			}

			err = <-done
			if err != nil {
				log.Errorf("failed to listen for deposit asset: %v\n", err)
				errOut <- err
				return
			}

			log.Infof("Asset successfully deposited with TX hash: %v\n", txHash.String())
		}

		// to common.Address, tokenId *big.Int, properties [][]byte, values [][]byte, salts [][32]byte
		args := []interface{}{requestData.To, requestData.TokenID, requestData.Props, requestData.Values, requestData.Salts}

		txID, done, err := s.identityService.Execute(ctx, req.RegistryAddress, GenericMintMethodABI, "mint", args...)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Sent off ethTX to mint [tokenID: %s, anchor: %s, nextAnchor: %s, registry: %s] to invoice unpaid contract.",
			hexutil.Encode(requestData.TokenID.Bytes()), hexutil.Encode(requestData.AnchorID[:]), hexutil.Encode(requestData.NextAnchorID.Bytes()), requestData.To.String())

		log.Debugf("To: %s", requestData.To.String())
		log.Debugf("TokenID: %s", hexutil.Encode(requestData.TokenID.Bytes()))
		log.Debugf("AnchorID: %s", hexutil.Encode(requestData.AnchorID[:]))
		log.Debugf("NextAnchorID: %s", hexutil.Encode(requestData.NextAnchorID.Bytes()))
		log.Debugf("Props: %s", byteSlicetoString(requestData.Props))
		log.Debugf("Values: %s", byteSlicetoString(requestData.Values))
		log.Debugf("Salts: %s", byte32SlicetoString(requestData.Salts))
		log.Debugf("Proofs: %s", byteByte32SlicetoString(requestData.Proofs))
		log.Debugf("Asset: %s", hexutil.Encode(requestData.BundledHash[:]))
		log.Debugf("AssetManager: %s", hexutil.Encode(req.AssetManagerAddress.Bytes()))

		err = <-done
		if err != nil {
			// some problem occurred in a child task
			errOut <- errors.New("mint nft failed for document %s and transaction %s with error %s", hexutil.Encode(req.DocumentID), txID, err.Error())
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

		log.Infof("Document %s minted successfully within transaction %s", hexutil.Encode(req.DocumentID), txID)

		errOut <- nil
		return
	}
}

func (s *service) transferFromJob(ctx context.Context, registry common.Address, from common.Address, to common.Address, tokenID TokenID) func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
		owner, err := s.OwnerOf(registry, tokenID[:])
		if err != nil {
			errOut <- errors.New("error while checking new NFT owner %v", err)
			return
		}
		if owner.Hex() != from.Hex() {
			errOut <- errors.New("from address is not the owner of tokenID %s from should be %s, instead got %s", tokenID.String(), from.Hex(), owner.Hex())
			return
		}

		txID, done, err := s.identityService.Execute(ctx, registry, ABI, "transferFrom", from, to, utils.ByteSliceToBigInt(tokenID[:]))
		if err != nil {
			errOut <- err
			return
		}
		log.Infof("sent off ethTX to transferFrom [registry: %s tokenID: %s, from: %s, to: %s].",
			registry.String(), tokenID.String(), from.String(), to.String())

		err = <-done
		if err != nil {
			// some problem occurred in a child task
			errOut <- errors.New("failed to transfer token with transaction:  %s with error %s", txID, err.Error())
			return
		}

		// Check if tokenID is new owner is to address
		owner, err = s.OwnerOf(registry, tokenID[:])
		if err != nil {
			errOut <- errors.New("error while checking new NFT owner %v", err)
			return
		}
		if owner.Hex() != to.Hex() {
			errOut <- errors.New("new owner for tokenID %s should be %s, instead got %s", tokenID.String(), registry.Hex(), owner.Hex())
			return
		}

		log.Infof("token %s successfully transferred from %s to %s with transaction %s ", tokenID.String(), from.Hex(), to.Hex(), txID)

		errOut <- nil
		return
	}
}

// OwnerOf returns the owner of the NFT token on ethereum chain
func (s *service) OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error) {
	c := s.bindCallerContract(registry, nftABI, s.ethClient)
	opts, cancF := s.ethClient.GetGethCallOpts(false)
	defer cancF()
	return owner, c.Call(opts, &owner, "ownerOf", tokenID)
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
