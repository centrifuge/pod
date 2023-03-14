package centchain

import (
	"context"
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/client"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/jobs"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

	ErrExtrinsicSubmission = errors.Error("couldn't submit extrinsic")

	ErrMultisigNotSupported = errors.Error("multi signature not supported")
)

func init() {
	gob.Register(ExtrinsicInfo{})
}

var log = logging.Logger("centchain-client")

type CallProviderFn func(metadata *types.Metadata) (*types.Call, error)

// ExtrinsicInfo holds details of a successful extrinsic
type ExtrinsicInfo struct {
	Hash      types.Hash
	BlockHash types.Hash
	Index     uint // index number of extrinsic in a block

	// EventsRaw contains all the events in the given block
	// if you want to filter events for an extrinsic, use the Index
	EventsRaw types.EventRecordsRaw
}

// Events returns all the events occurred in a given block
func (e ExtrinsicInfo) Events(meta *types.Metadata) (events Events, err error) {
	return events, e.EventsRaw.DecodeEventRecords(meta, &events)
}

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

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
	SubmitAndWatch(
		ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (ExtrinsicInfo, error)

	// GetStorageLatest returns the latest value at the given key
	GetStorageLatest(key types.StorageKey, target interface{}) (bool, error)

	// GetBlockLatest returns the latest block
	GetBlockLatest() (*types.SignedBlock, error)

	// GetBlockHash returns the hash of a block
	GetBlockHash(blockNumber uint64) (types.Hash, error)

	// GetBlock returns the block
	GetBlock(blockHash types.Hash) (*types.SignedBlock, error)

	// GetPendingExtrinsics returns all pending extrinsics
	GetPendingExtrinsics() ([]types.Extrinsic, error)
}

//go:generate mockery --name substrateAPI --structname SubstrateAPIMock --filename substrate_api_mock.go --inpackage

// substrateAPI exposes Substrate API functions
type substrateAPI interface {
	GetMetadataLatest() (*types.Metadata, error)
	Call(result interface{}, method string, args ...interface{}) error
	GetBlockHash(blockNumber uint64) (types.Hash, error)
	GetBlockLatest() (*types.SignedBlock, error)
	GetRuntimeVersionLatest() (*types.RuntimeVersion, error)
	SubmitExtrinsic(ext types.Extrinsic) (types.Hash, error)
	GetStorageLatest(key types.StorageKey, target interface{}) (bool, error)
	GetStorage(key types.StorageKey, target interface{}, blockHash types.Hash) error
	GetBlock(blockHash types.Hash) (*types.SignedBlock, error)
	GetPendingExtrinsics() ([]types.Extrinsic, error)
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

func (dsa *defaultSubstrateAPI) SubmitExtrinsic(ext types.Extrinsic) (types.Hash, error) {
	return dsa.sapi.RPC.Author.SubmitExtrinsic(ext)
}

func (dsa *defaultSubstrateAPI) GetStorageLatest(key types.StorageKey, target interface{}) (bool, error) {
	return dsa.sapi.RPC.State.GetStorageLatest(key, target)
}

func (dsa *defaultSubstrateAPI) GetPendingExtrinsics() ([]types.Extrinsic, error) {
	return dsa.sapi.RPC.Author.PendingExtrinsics()
}

type api struct {
	sapi       substrateAPI
	dispatcher jobs.Dispatcher
	accounts   map[string]uint32
	accMu      sync.Mutex // accMu to protect accounts

	centChainMaxRetries    int
	centChainRetryInterval time.Duration
}

// NewAPI returns a new centrifuge chain api.
func NewAPI(
	sapi substrateAPI,
	dispatcher jobs.Dispatcher,
	centChainMaxRetries int,
	centChainRetryInterval time.Duration,
) API {
	return &api{
		sapi:                   sapi,
		dispatcher:             dispatcher,
		accounts:               make(map[string]uint32),
		accMu:                  sync.Mutex{},
		centChainMaxRetries:    centChainMaxRetries,
		centChainRetryInterval: centChainRetryInterval,
	}
}

func (a *api) Call(result interface{}, method string, args ...interface{}) error {
	return a.sapi.Call(result, method, args...)
}

func (a *api) GetMetadataLatest() (*types.Metadata, error) {
	return a.sapi.GetMetadataLatest()
}

func (a *api) GetStorageLatest(key types.StorageKey, target interface{}) (bool, error) {
	return a.sapi.GetStorageLatest(key, target)
}

func (a *api) submitExtrinsic(
	c types.Call,
	nonce uint64,
	krp signature.KeyringPair,
) (
	txHash types.Hash,
	bn types.BlockNumber,
	sig types.MultiSignature,
	err error,
) {
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

	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                era,
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(nonce),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(krp, o)
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlock, err := a.sapi.GetBlockLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlockNumber := startBlock.Block.Header.Number
	txHash, err = a.sapi.SubmitExtrinsic(ext)
	return txHash, startBlockNumber, ext.Signature.Signature, err
}

func (a *api) SubmitExtrinsic(_ context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (types.Hash, types.BlockNumber, types.MultiSignature, error) {
	var current int
	var err error
	var txHash types.Hash
	var bn types.BlockNumber
	var sig types.MultiSignature
	var nonce uint32

	for {
		if current >= a.centChainMaxRetries {
			err = errors.Error("max concurrent transaction tries reached")
			return types.Hash{}, types.BlockNumber(0), types.MultiSignature{}, errors.NewTypedError(ErrCentChainTransaction, err)
		}

		current++
		var ok bool
		nonce, ok = a.getNonceInAccount(krp.PublicKey)
		if !ok { // First time using account in session
			nonce, err = a.getNonceFromChain(meta, krp.PublicKey)
			if err != nil {
				return txHash, bn, sig, err
			}
			a.setNonceInAccount(krp.PublicKey, nonce)
		}

		txHash, bn, sig, err = a.submitExtrinsic(c, uint64(nonce), krp)
		if err == nil {
			break
		}

		if strings.Contains(err.Error(), ErrNonceTooLow.Error()) || strings.Contains(err.Error(), ErrInvalidTransaction.Error()) {
			log.Warnf("Used Nonce %v. Failed with error: %v\n", nonce, err)
			log.Warnf("Concurrent transaction identified, trying again [%d/%d]\n", current, a.centChainMaxRetries)
			chainNonce, err := a.getNonceFromChain(meta, krp.PublicKey)
			if err != nil {
				return txHash, bn, sig, err
			}
			a.setNonceInAccount(krp.PublicKey, chainNonce)

			time.Sleep(a.centChainRetryInterval)

			continue
		}

		return txHash, bn, sig, err
	}

	encodedCall, _ := codec.Encode(c)

	log.Infof("Successfully submitted ext %s for call %s with nonce %d and from blockNumber %d", txHash.Hex(), hexutil.Encode(encodedCall), nonce, bn)
	a.incrementNonce(krp.PublicKey)
	return txHash, bn, sig, nil
}

// SubmitAndWatch is submitting a CentChain transaction and starts a task to wait for the transaction result
func (a *api) SubmitAndWatch(
	ctx context.Context,
	meta *types.Metadata,
	c types.Call,
	krp signature.KeyringPair,
) (info ExtrinsicInfo, err error) {
	identity, err := contextutil.Identity(ctx)
	if err != nil {
		return info, errors.ErrContextIdentityRetrieval
	}

	txHash, bn, sig, err := a.SubmitExtrinsic(ctx, meta, c, krp)
	if err != nil {
		return info, ErrExtrinsicSubmission
	}

	s, err := getSignature(sig)
	if err != nil {
		return info, err
	}

	task := getTaskName(txHash)

	a.dispatcher.RegisterRunnerFunc(task, a.getDispatcherRunnerFunc(&bn, txHash, s, meta))

	job := gocelery.NewRunnerFuncJob("", task, nil, nil, time.Time{})
	res, err := a.dispatcher.Dispatch(identity, job)
	if err != nil {
		return info, fmt.Errorf("failed to dispatch job: %w", err)
	}

	result, err := res.Await(context.Background())
	if err != nil {
		return info, err
	}

	return result.(ExtrinsicInfo), nil
}

func (a *api) GetBlockLatest() (*types.SignedBlock, error) {
	return a.sapi.GetBlockLatest()
}

func (a *api) GetBlockHash(blockNumber uint64) (types.Hash, error) {
	return a.sapi.GetBlockHash(blockNumber)
}

func (a *api) GetBlock(blockHash types.Hash) (*types.SignedBlock, error) {
	return a.sapi.GetBlock(blockHash)
}

func (a *api) GetPendingExtrinsics() ([]types.Extrinsic, error) {
	return a.sapi.GetPendingExtrinsics()
}

func (a *api) getDispatcherRunnerFunc(
	blockNumber *types.BlockNumber,
	txHash types.Hash,
	sig types.Signature,
	meta *types.Metadata,
) gocelery.RunnerFunc {
	fn := func(_ []interface{}, _ map[string]interface{}) (interface{}, error) {
		bh, err := a.sapi.GetBlockHash(uint64(*blockNumber))
		if err != nil {
			return nil, fmt.Errorf("failed to get block hash for block number %d: %w", *blockNumber, err)
		}

		block, err := a.sapi.GetBlock(bh)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %d: %w", *blockNumber, err)
		}

		extIdx := isExtrinsicSignatureInBlock(sig, block.Block)
		if extIdx < 0 {
			log.Debugf("Extrinsic %s not found in block %d, trying in next block...", txHash.Hex(), *blockNumber)
			*blockNumber++
			return nil, fmt.Errorf("extrinsic %s not found in block %d", txHash.Hex(), *blockNumber)
		}

		log.Debugf("Extrinsic %s found in block %d", txHash.Hex(), *blockNumber)

		eventsRaw, err := a.checkExtrinsicEventSuccess(meta, bh, extIdx)
		if err != nil {
			return nil, err
		}

		info := ExtrinsicInfo{
			Hash:      txHash,
			BlockHash: bh,
			Index:     uint(extIdx),
			EventsRaw: eventsRaw,
		}

		return info, nil
	}

	return fn
}

func (a *api) checkExtrinsicEventSuccess(
	meta *types.Metadata,
	blockHash types.Hash,
	extrinsicIdx int,
) (eventsRaw types.EventRecordsRaw, err error) {
	key, err := types.CreateStorageKey(meta, "System", "Events")
	if err != nil {
		return nil, err
	}

	err = a.sapi.GetStorage(key, &eventsRaw, blockHash)
	if err != nil {
		return nil, err
	}

	events := Events{}
	err = eventsRaw.DecodeEventRecords(meta, &events)
	if err != nil {
		return nil, err
	}

	// Check success events
	for _, es := range events.System_ExtrinsicSuccess {
		if es.Phase.IsApplyExtrinsic && es.Phase.AsApplyExtrinsic == uint32(extrinsicIdx) {
			// Extra check for proxy calls.
			if err := checkSuccessfulProxyExecution(events, meta, extrinsicIdx); err != nil {
				return nil, errors.New("proxy call was not successful: %s", err)
			}

			return eventsRaw, nil
		}
	}

	// Otherwise, check failure events
	for _, es := range events.System_ExtrinsicFailed {
		if es.Phase.IsApplyExtrinsic && es.Phase.AsApplyExtrinsic == uint32(extrinsicIdx) {
			return nil, handleDispatchError(meta, es.DispatchError, extrinsicIdx)
		}
	}

	return nil, errors.New("should not have reached this step: %v", events)
}

func handleDispatchError(meta *types.Metadata, dispatchError types.DispatchError, extrinsicIdx int) error {
	if dispatchError.IsModule {
		moduleErr := dispatchError.ModuleError
		if metaErr, findErr := meta.FindError(moduleErr.Index, moduleErr.Error); findErr == nil {
			return errors.New(
				"extrinsic %d failed with '%s - %s'",
				extrinsicIdx,
				metaErr.Name,
				metaErr.Value,
			)
		}
	}

	return errors.New("extrinsic %d failed: %v", extrinsicIdx, dispatchError)
}

func checkSuccessfulProxyExecution(events Events, meta *types.Metadata, extrinsicIdx int) error {
	for _, event := range events.Proxy_ProxyExecuted {
		if event.Phase.IsApplyExtrinsic && event.Phase.AsApplyExtrinsic == uint32(extrinsicIdx) {
			if !event.Result.Ok {
				return handleDispatchError(meta, event.Result.Error, extrinsicIdx)
			}

			return nil
		}
	}

	return nil
}

const (
	taskNameFormat = "cent_chain_tx_status-%s"
)

func getTaskName(hash types.Hash) string {
	return fmt.Sprintf(taskNameFormat, hash.Hex())
}

func (a *api) incrementNonce(accountID []byte) {
	a.accMu.Lock()
	defer a.accMu.Unlock()
	acc := hexutil.Encode(accountID)
	if _, ok := a.accounts[acc]; !ok { // Should not be reached
		return
	}
	a.accounts[acc]++
}

func (a *api) getNonceFromChain(meta *types.Metadata, krp []byte) (uint32, error) {
	key, err := types.CreateStorageKey(meta, "System", "Account", krp, nil)
	if err != nil {
		return 0, err
	}

	var accountInfo types.AccountInfo

	ok, err := a.sapi.GetStorageLatest(key, &accountInfo)

	if err != nil {
		return 0, err
	}

	if !ok {
		return 0, errors.New("account information not found on chain")
	}

	return uint32(accountInfo.Nonce), nil
}

func (a *api) setNonceInAccount(accountID []byte, nonce uint32) {
	a.accMu.Lock()
	defer a.accMu.Unlock()

	a.accounts[hexutil.Encode(accountID)] = nonce
}

func (a *api) getNonceInAccount(accountID []byte) (uint32, bool) {
	a.accMu.Lock()
	defer a.accMu.Unlock()

	n, ok := a.accounts[hexutil.Encode(accountID)]
	return n, ok
}

func getSignature(msig types.MultiSignature) (types.Signature, error) {
	if msig.IsEd25519 {
		return msig.AsEd25519, nil
	}
	if msig.IsSr25519 {
		return msig.AsSr25519, nil
	}
	return types.Signature{}, ErrMultisigNotSupported
}

func isExtrinsicSignatureInBlock(extSign types.Signature, block types.Block) int {
	for idx, xx := range block.Extrinsics {
		switch {
		case xx.Signature.Signature.IsSr25519 && xx.Signature.Signature.AsSr25519 == extSign:
			return idx
		case xx.Signature.Signature.IsEd25519 && xx.Signature.Signature.AsEd25519 == extSign:
			return idx
		}
	}

	return -1
}
