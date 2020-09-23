package oracle

import (
	"context"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("oracle")

// oracleABI is the default abi for caller functions on the NFT oracle
var oracleABI abi.ABI

const (
	// ErrFingerprintMismatch
	ErrFingerprintMismatch = errors.Error("Fingerprint mismatch")

	// NFTValueUpdatedSignature used for finding events
	NFTValueUpdatedSignature = "NFTValueUpdated(uint indexed)"

	// GenericMintMethodABI constant interface to interact with mint methods
	// TODO: update with update function
	UpdateMethodABI = ``

	// ABI is string abi with required methods to call the NFT oracle
	// TODO: paste in oracle contract abi
	ABI = ``
)

func init() {
	var err error
	oracleABI, err = abi.JSON(strings.NewReader(ABI))
	if err != nil {
		log.Fatalf("failed to decode NFT ABI: %v", err)
	}
}

// Config is the config interface for nft package
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
}

// service handles all interactions related to minting of NFTs for unpaid invoices on Ethereum
type service struct {
	cfg                Config
	identityService    identity.Service
	ethClient          ethereum.Client
	queue              queue.TaskQueuer
	bindCallerContract func(address common.Address, abi abi.ABI, client ethereum.Client) *bind.BoundContract
	jobsManager        jobs.Manager
	blockHeightFunc    func() (height uint64, err error)
}

// newService creates the NFT Oracle Service given the parameters
func newService(
	cfg Config,
	identityService identity.Service,
	ethClient ethereum.Client,
	queue queue.TaskQueuer,
	bindCallerContract func(address common.Address, abi abi.ABI, client ethereum.Client) *bind.BoundContract,
	jobsMan jobs.Manager,
	blockHeightFunc func() (uint64, error)) *service {
	return &service{
		cfg:                cfg,
		identityService:    identityService,
		ethClient:          ethClient,
		queue:              queue,
		bindCallerContract: bindCallerContract,
		jobsManager:        jobsMan,
		blockHeightFunc:    blockHeightFunc,
	}
}

// Update NFT Oracle updates the NFT Oracle contract with the risk and value of an NFT
func (s *service) UpdateNFTOracle(ctx context.Context, req updateNFTOracleRequest) (*UpdateResponse, chan error, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	didBytes := tc.GetIdentityID()

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, nil, err
	}

	jobID, done, err := s.jobsManager.ExecuteWithinJob(contextutil.Copy(ctx), did, jobs.NilJobID(), "Updating NFT Oracle",
		s.updateOracleJob(ctx, req))

	if err != nil {
		return nil, nil, err
	}

	return &UpdateResponse{
		JobID: jobID.String(),
	}, done, nil
}

func (s *service) updateOracleJob(ctx context.Context, req updateNFTOracleRequest) func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
		// to common.Address, tokenId *big.Int, bytes32, properties [][]byte, values [][]byte, salts [][32]byte
		args := []interface{}{req.TokenID, req.OracleFingerprint, req.Result}

		block, err := s.ethClient.GetEthClient().BlockByNumber(context.Background(), nil)
		if err != nil {
			errOut <- errors.New("failed to get latest block: %v", err)
			return
		}

		txID, done, err := s.identityService.Execute(ctx, req.OracleAddress, UpdateMethodABI, "update", args...)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Sent off ethTX to update NFT oracle[tokenID: %s, nftOracleAddress: %s, request: %s] to NFT Oracle contract.",
			hexutil.Encode(req.TokenID[:]),
			req.OracleAddress.String(),
			hexutil.Encode(req.Result))

		err = <-done
		if err != nil {
			// some problem occurred in a child task
			errOut <- errors.New("update nft oracle contract failed for tokenID %s and transaction %s with error %s", hexutil.Encode(req.TokenID[:]), txID, err.Error())
			return
		}

		txHash, done, err := ethereum.CreateWaitForEventJob(
			ctx, txMan, s.queue, accountID, jobID,
			// TODO: HOW TO GET THE TOPIC HASH HERE?
			NFTValueUpdatedSignature, block.Number(), req.OracleAddress, requestData.BundledHash)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Asset successfully deposited with TX hash: %v\n", txHash.String())

		errOut <- nil
	}
}
