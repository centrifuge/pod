package centchain

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/client"
	"github.com/centrifuge/go-substrate-rpc-client/rpc/author"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	logging "github.com/ipfs/go-log"
)

const (
	// ErrCentChainTransaction is a generic error type to be used for CentChain errors
	ErrCentChainTransaction = errors.Error("error on centchain tx layer")

	// ErrNonceTooLow nonce is too low
	ErrNonceTooLow = errors.Error("Priority is too low")

	// ErrInvalidTransaction wrapper for a general error
	// Used sometimes as stale extrinsic (nonce too low)
	ErrInvalidTransaction = errors.Error("Invalid Transaction")
)

var log = logging.Logger("centchain-client")

// API exposes required functions to interact with Centrifuge Chain.
type API interface {

	// Call allows to make a read operation
	Call(result interface{}, method string, args ...interface{}) error

	// GetMetadataLatest returns latest metadata from the centrifuge chain.
	GetMetadataLatest() (*types.Metadata, error)

	// SubmitExtrinsic signs the given call with the provided KeyRingPair and submits an extrinsic.
	// Returns transaction hash, latest block number before extrinsic submission, and signature attached with the extrinsic.
	SubmitExtrinsic(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.MultiSignature, err error)

	// SubmitAndWatch returns function that submits and watches an extrinsic, implements transaction.Submitter
	SubmitAndWatch(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) func(accountID identity.DID, jobID jobs.JobID, jobMan jobs.Manager, errOut chan<- error)
}

// SubstrateAPI exposes Substrate API functions
type SubstrateAPI interface {
	GetMetadataLatest() (*types.Metadata, error)
	Call(result interface{}, method string, args ...interface{}) error
	GetBlockHash(blockNumber uint64) (types.Hash, error)
	GetBlockLatest() (*types.SignedBlock, error)
	GetRuntimeVersionLatest() (*types.RuntimeVersion, error)
	GetClient() client.Client
	GetStorageLatest(key types.StorageKey, target interface{}) error
	GetStorage(key types.StorageKey, target interface{}, blockHash types.Hash) error
	GetBlock(blockHash types.Hash) (*types.SignedBlock, error)
}

// Config defines functions to get centchain details
type Config interface {
	GetCentChainIntervalRetry() time.Duration
	GetCentChainMaxRetries() int
	GetCentChainAccount() (acc config.CentChainAccount, err error)
}

type defaultSubstrateAPI struct {
	sapi *gsrpc.SubstrateAPI
}

func (dsa *defaultSubstrateAPI) GetMetadataLatest() (*types.Metadata, error) {
	return dsa.sapi.RPC.State.GetMetadataLatest()
}

func (dsa *defaultSubstrateAPI) Call(result interface{}, method string, args ...interface{}) error {
	return dsa.sapi.Client.Call(result, method, args...)
}

func (dsa *defaultSubstrateAPI) GetBlockHash(blockNumber uint64) (types.Hash, error) {
	return dsa.sapi.RPC.Chain.GetBlockHash(blockNumber)
}

func (dsa *defaultSubstrateAPI) GetBlock(blockHash types.Hash) (*types.SignedBlock, error) {
	return dsa.sapi.RPC.Chain.GetBlock(blockHash)
}

func (dsa *defaultSubstrateAPI) GetStorage(key types.StorageKey, target interface{}, blockHash types.Hash) error {
	_, err := dsa.sapi.RPC.State.GetStorage(key, target, blockHash)
	return err
}

func (dsa *defaultSubstrateAPI) GetBlockLatest() (*types.SignedBlock, error) {
	return dsa.sapi.RPC.Chain.GetBlockLatest()
}

func (dsa *defaultSubstrateAPI) GetRuntimeVersionLatest() (*types.RuntimeVersion, error) {
	return dsa.sapi.RPC.State.GetRuntimeVersionLatest()
}

func (dsa *defaultSubstrateAPI) GetClient() client.Client {
	return dsa.sapi.Client
}

func (dsa *defaultSubstrateAPI) GetStorageLatest(key types.StorageKey, target interface{}) error {
	_, err := dsa.sapi.RPC.State.GetStorageLatest(key, target)
	return err
}

type api struct {
	sapi     SubstrateAPI
	config   Config
	queueSrv *queue.Server
	accounts map[string]uint32
	accMu    sync.Mutex // accMu to protect accounts
	mu       sync.Mutex
}

// NewAPI returns a new centrifuge chain api.
func NewAPI(sapi SubstrateAPI, config Config, queueSrv *queue.Server) API {
	return &api{
		sapi:     sapi,
		config:   config,
		queueSrv: queueSrv,
		accounts: map[string]uint32{},
		accMu:    sync.Mutex{},
		mu:       sync.Mutex{},
	}
}

func (a *api) Call(result interface{}, method string, args ...interface{}) error {
	return a.sapi.Call(result, method, args...)
}

func (a *api) GetMetadataLatest() (*types.Metadata, error) {
	return a.sapi.GetMetadataLatest()
}

func (a *api) SubmitExtrinsic(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.MultiSignature, err error) {
	ext := types.NewExtrinsic(c)
	era := types.ExtrinsicEra{IsMortalEra: false}

	genesisHash, err := a.sapi.GetBlockHash(0)
	if err != nil {
		return txHash, bn, sig, err
	}

	rv, err := a.sapi.GetRuntimeVersionLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	nonce, err := contextutil.Nonce(ctx)
	if err != nil {
		return txHash, bn, sig, err
	}

	o := types.SignatureOptions{
		BlockHash:   genesisHash,
		Era:         era,
		GenesisHash: genesisHash,
		Nonce:       types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion: rv.SpecVersion,
		Tip:         types.NewUCompactFromUInt(0),
		TransactionVersion: 1,
	}

	err = ext.Sign(krp, o)
	if err != nil {
		return txHash, bn, sig, err
	}

	auth := author.NewAuthor(a.sapi.GetClient())
	startBlock, err := a.sapi.GetBlockLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlockNumber := startBlock.Block.Header.Number
	txHash, err = auth.SubmitExtrinsic(ext)
	return txHash, startBlockNumber, ext.Signature.Signature, err
}

func (a *api) QueueCentChainEXTStatusTask(
	accountID identity.DID,
	jobID jobs.JobID,
	txHash types.Hash,
	fromBlock uint32,
	sig types.Signature,
	queuer queue.TaskQueuer) (res queue.TaskResult, err error) {

	params := map[string]interface{}{
		jobs.JobIDParam:              jobID.String(),
		TransactionAccountParam:      accountID.String(),
		TransactionExtHashParam:      txHash.Hex(),
		TransactionFromBlockParam:    fromBlock,
		TransactionExtSignatureParam: sig.Hex(),
	}

	return queuer.EnqueueJob(ExtrinsicStatusTaskName, params)
}

/**
SubmitWithRetries submits extrinsic to the centchain
Blocking Function that sends Extrinsic wrapped in a retrial block. It is based on the ErrNonceTooLow error,
meaning that a transaction is being attempted to run twice with the same nonce.
*/
func (a *api) SubmitWithRetries(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (types.Hash, types.BlockNumber, types.MultiSignature, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var current int
	var err error
	var txHash types.Hash
	var bn types.BlockNumber
	var sig types.MultiSignature
	var nonce uint32

	maxTries := a.config.GetCentChainMaxRetries()
	defaultAccount, err := a.config.GetCentChainAccount()
	if err != nil {
		return txHash, bn, sig, err
	}

	for {
		if current >= maxTries {
			err = errors.Error("max concurrent transaction tries reached")
			log.Errorf(err.Error())
			return types.Hash{}, types.BlockNumber(0), types.MultiSignature{}, errors.NewTypedError(ErrCentChainTransaction, err)
		}

		current++

		var ok bool
		nonce, ok = a.getNonceInAccount(defaultAccount.ID)
		if !ok { // First time using account in session
			nonce, err = a.getNonceFromChain(meta, krp.PublicKey)
			if err != nil {
				return txHash, bn, sig, err
			}
			a.setNonceInAccount(defaultAccount.ID, nonce)
		}
		txHash, bn, sig, err = a.SubmitExtrinsic(contextutil.WithNonce(ctx, nonce), meta, c, krp)
		if err == nil {
			break
		}

		if strings.Contains(err.Error(), ErrNonceTooLow.Error()) || strings.Contains(err.Error(), ErrInvalidTransaction.Error()) {
			log.Warningf("Used Nonce %v. Failed with error: %v\n", nonce, err)
			log.Warningf("Concurrent transaction identified, trying again [%d/%d]\n", current, maxTries)
			chainNonce, err := a.getNonceFromChain(meta, krp.PublicKey)
			if err != nil {
				return txHash, bn, sig, err
			}
			a.setNonceInAccount(defaultAccount.ID, chainNonce)
			time.Sleep(a.config.GetCentChainIntervalRetry())
			continue
		}

		return txHash, bn, sig, err
	}

	log.Infof("Successfully submitted ext %s with nonce %d and from blockNumber %d", txHash.Hex(), nonce, bn)
	a.incrementNonce(defaultAccount.ID)

	return txHash, bn, sig, nil
}

// SubmitAndWatch is submitting a CentChain transaction and starts a task to wait for the transaction result
func (a *api) SubmitAndWatch(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) func(accountID identity.DID, jobID jobs.JobID, jobsMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, jobMan jobs.Manager, errOut chan<- error) {
		tx, bn, msig, err := a.SubmitWithRetries(ctx, meta, c, krp)
		if err != nil {
			errOut <- err
			return
		}

		sig, err := getSignature(msig)
		if err != nil {
			errOut <- err
			return
		}

		res, err := a.QueueCentChainEXTStatusTask(accountID, jobID, tx, uint32(bn), sig, a.queueSrv)
		if err != nil {
			errOut <- err
			return
		}

		_, err = res.Get(jobMan.GetDefaultTaskTimeout())
		if err != nil {
			errOut <- err
			return
		}
		errOut <- nil
	}
}

func (a *api) incrementNonce(accountID string) {
	a.accMu.Lock()
	defer a.accMu.Unlock()
	if _, ok := a.accounts[accountID]; !ok { // Should not be reached
		return
	}
	a.accounts[accountID] = a.accounts[accountID] + 1
}

func (a *api) getNonceFromChain(meta *types.Metadata, krp []byte) (uint32, error) {
	key, err := types.CreateStorageKey(meta, "System", "Account", krp, nil)
	if err != nil {
		return 0, err
	}

	var accountInfo types.AccountInfo
	err = a.sapi.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return 0, err
	}
	return uint32(accountInfo.Nonce), nil
}

func (a *api) setNonceInAccount(accountID string, nonce uint32) {
	a.accMu.Lock()
	defer a.accMu.Unlock()

	a.accounts[accountID] = nonce
}

func (a *api) getNonceInAccount(accountID string) (uint32, bool) {
	a.accMu.Lock()
	defer a.accMu.Unlock()

	n, ok := a.accounts[accountID]
	return n, ok
}

func getSignature(msig types.MultiSignature) (types.Signature, error) {
	if msig.IsEd25519 {
		return msig.AsEd25519, nil
	}
	if msig.IsSr25519 {
		return msig.AsSr25519, nil
	}
	return types.Signature{}, errors.New("MultiSignature not supported")
}
