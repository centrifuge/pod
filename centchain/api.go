package centchain

import (
	"reflect"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	logging "github.com/ipfs/go-log"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/client"
	"github.com/centrifuge/go-substrate-rpc-client/rpc/author"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

const (
	// ErrCentChainTransaction is a generic error type to be used for CentChain errors
	ErrCentChainTransaction = errors.Error("error on centchain tx layer")

	// ErrTransactionUnderpriced transaction is under priced
	ErrTransactionUnderpriced = errors.Error("replacement transaction underpriced")

	// ErrNonceTooLow nonce is too low
	ErrNonceTooLow = errors.Error("nonce too low")
)

var log = logging.Logger("centchain-client")

// API exposes required functions to interact with Centrifuge Chain.
type API interface {
	// GetMetadataLatest returns latest metadata from the centrifuge chain.
	GetMetadataLatest() (*types.Metadata, error)

	// SubmitExtrinsic signs the given call with the provided KeyRingPair and submits an extrinsic.
	// Returns transaction hash, latest block number before extrinsic submission, and signature attached with the extrinsic.
	SubmitExtrinsic(meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error)
}

type Config interface {
	GetCentChainIntervalRetry() time.Duration
	GetCentChainMaxRetries() int
}

type api struct {
	getBlockHash            func(uint64) (types.Hash, error)
	getRuntimeVersionLatest func() (*types.RuntimeVersion, error)
	getStorageLatest        func(key types.StorageKey, target interface{}) error
	getClient               func() client.Client
	getBlockLatest          func() (*types.SignedBlock, error)
	getMetadataLatest       func() (*types.Metadata, error)
	config                  Config
	mu                      sync.Mutex
}

// NewAPI returns a new centrifuge chain api.
func NewAPI(sapi *gsrpc.SubstrateAPI) API {
	return api{
		getBlockHash:            sapi.RPC.Chain.GetBlockHash,
		getRuntimeVersionLatest: sapi.RPC.State.GetRuntimeVersionLatest,
		getStorageLatest:        sapi.RPC.State.GetStorageLatest,
		getClient:               func() client.Client { return sapi.Client },
		getBlockLatest:          sapi.RPC.Chain.GetBlockLatest,
		getMetadataLatest:       sapi.RPC.State.GetMetadataLatest,
	}
}

func (a api) GetMetadataLatest() (*types.Metadata, error) {
	return a.getMetadataLatest()
}

func (a api) SubmitExtrinsic(meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ext := types.NewExtrinsic(c)
	era := types.ExtrinsicEra{IsMortalEra: false}

	genesisHash, err := a.getBlockHash(0)
	if err != nil {
		return txHash, bn, sig, err
	}

	rv, err := a.getRuntimeVersionLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	key, err := types.CreateStorageKey(meta, "System", "AccountNonce", krp.PublicKey)
	if err != nil {
		return txHash, bn, sig, err
	}

	var nonce uint32
	err = a.getStorageLatest(key, &nonce)
	if err != nil {
		return txHash, bn, sig, err
	}

	o := types.SignatureOptions{
		BlockHash:   genesisHash,
		Era:         era,
		GenesisHash: genesisHash,
		Nonce:       types.UCompact(nonce),
		SpecVersion: rv.SpecVersion,
		Tip:         0,
	}

	err = ext.Sign(krp, o)
	if err != nil {
		return txHash, bn, sig, err
	}

	auth := author.NewAuthor(a.getClient())
	startBlock, err := a.getBlockLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlockNumber := startBlock.Block.Header.Number
	txHash, err = auth.SubmitExtrinsic(ext)
	return txHash, startBlockNumber, ext.Signature.Signature, err
}

func (a api) QueueCentChainEXTStatusTask(
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

	return queuer.EnqueueJobWithMaxTries(ExtrinsicStatusTaskName, params)
}

/**
SubmitWithRetries submits transaction to the centchain
Blocking Function that sends transaction using reflection wrapped in a retrial block. It is based on the ErrTransactionUnderpriced error,
meaning that a transaction is being attempted to run twice, and the logic is to override the existing one. As we have constant
gas prices that means that a concurrent transaction race condition event has happened.
- method: Transaction Method
- params: Arbitrary number of parameters that are passed to the function fname call
Note: method must always return "txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error"
*/
func (a api) SubmitWithRetries(method interface{}, params ...interface{}) (types.Hash, types.BlockNumber, types.Signature, error) {
	f := reflect.ValueOf(method)
	maxTries := a.config.GetCentChainMaxRetries()

	var current int
	var err error
	for {
		if current >= maxTries {
			return types.Hash{}, types.BlockNumber(0), types.Signature{}, errors.NewTypedError(ErrCentChainTransaction, errors.New("max concurrent transaction tries reached: %v", err))
		}

		var in []reflect.Value
		for _, p := range params {
			in = append(in, reflect.ValueOf(p))
		}

		var txHash types.Hash
		var bn types.BlockNumber
		var sig types.Signature

		result := f.Call(in)
		if !result[0].IsNil() {
			txHash = result[0].Interface().(types.Hash)
		}

		if !result[1].IsNil() {
			bn = result[1].Interface().(types.BlockNumber)
		}

		if !result[2].IsNil() {
			sig = result[2].Interface().(types.Signature)
		}

		if !result[3].IsNil() {
			err = result[3].Interface().(error)
		}

		if err == nil {
			return txHash, bn, sig, nil
		}

		if (err.Error() == ErrTransactionUnderpriced.Error()) || (err.Error() == ErrNonceTooLow.Error()) {
			log.Warningf("Concurrent transaction identified, trying again [%d/%d]\n", current, maxTries)
			time.Sleep(a.config.GetCentChainIntervalRetry())
			continue
		}

		return types.Hash{}, types.BlockNumber(0), types.Signature{}, err

	}
}

// SubmitAndWatch is submitting a CentChain transaction and starts a task to wait for the transaction result
func (a api) SubmitAndWatch(method interface{}, params ...interface{}) func(accountID identity.DID, jobID jobs.JobID, jobsMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, jobMan jobs.Manager, errOut chan<- error) {
		tx, bn, sig, err := a.SubmitWithRetries(method, params)
		if err != nil {
			errOut <- err
			return
		}

		res, err := a.QueueCentChainEXTStatusTask(accountID, jobID, tx, uint32(bn), sig, nil)
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
