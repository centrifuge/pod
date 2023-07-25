package centchain

import (
	"context"
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/registry/parser"

	"github.com/centrifuge/go-substrate-rpc-client/v4/registry"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/client"
	"github.com/centrifuge/go-substrate-rpc-client/v4/registry/retriever"
	"github.com/centrifuge/go-substrate-rpc-client/v4/rpc/author"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
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
)

func init() {
	gob.Register(ExtrinsicInfo{})
	gob.Register(registry.DecodedFields{})
}

var log = logging.Logger("centchain-client")

// ExtrinsicInfo holds details of a successful extrinsic
type ExtrinsicInfo struct {
	Hash      types.Hash
	BlockHash types.Hash
	Index     uint // index number of extrinsic in a block

	// EventsRaw contains all the events in the given block
	// if you want to filter events for an extrinsic, use the Index
	Events []*parser.Event
}

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

	// GetStorageLatest returns latest value at the given key
	GetStorageLatest(key types.StorageKey, target interface{}) error
}

// substrateAPI exposes Substrate API functions
type substrateAPI interface {
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
	sapi           substrateAPI
	config         Config
	dispatcher     jobs.Dispatcher
	accounts       map[string]uint32
	accMu          sync.Mutex // accMu to protect accounts
	eventRetriever retriever.EventRetriever
}

// NewAPI returns a new centrifuge chain api.
func NewAPI(sapi substrateAPI, config Config, dispatcher jobs.Dispatcher, eventRetriever retriever.EventRetriever) API {
	return &api{
		sapi:           sapi,
		config:         config,
		dispatcher:     dispatcher,
		accounts:       make(map[string]uint32),
		accMu:          sync.Mutex{},
		eventRetriever: eventRetriever,
	}
}

func (a *api) Call(result interface{}, method string, args ...interface{}) error {
	return a.sapi.Call(result, method, args...)
}

func (a *api) GetMetadataLatest() (*types.Metadata, error) {
	return a.sapi.GetMetadataLatest()
}

func (a *api) GetStorageLatest(key types.StorageKey, target interface{}) error {
	return a.sapi.GetStorageLatest(key, target)
}

func (a *api) submitExtrinsic(c types.Call, nonce uint64, krp signature.KeyringPair) (txHash types.Hash,
	bn types.BlockNumber, sig types.MultiSignature, err error) {
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

	auth := author.NewAuthor(a.sapi.GetClient())
	startBlock, err := a.sapi.GetBlockLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlockNumber := startBlock.Block.Header.Number
	txHash, err = auth.SubmitExtrinsic(ext)
	return txHash, startBlockNumber, ext.Signature.Signature, err
}

func (a *api) SubmitExtrinsic(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (types.Hash, types.BlockNumber, types.MultiSignature, error) {
	var current int
	var err error
	var txHash types.Hash
	var bn types.BlockNumber
	var sig types.MultiSignature
	var nonce uint32

	maxTries := a.config.GetCentChainMaxRetries()
	for {
		if current >= maxTries {
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
			log.Warnf("Concurrent transaction identified, trying again [%d/%d]\n", current, maxTries)
			chainNonce, err := a.getNonceFromChain(meta, krp.PublicKey)
			if err != nil {
				return txHash, bn, sig, err
			}
			a.setNonceInAccount(krp.PublicKey, chainNonce)
			time.Sleep(a.config.GetCentChainIntervalRetry())
			continue
		}

		return txHash, bn, sig, err
	}

	log.Infof("Successfully submitted ext %s with nonce %d and from blockNumber %d", txHash.Hex(), nonce, bn)
	a.incrementNonce(krp.PublicKey)
	return txHash, bn, sig, nil
}

// SubmitAndWatch is submitting a CentChain transaction and starts a task to wait for the transaction result
func (a *api) SubmitAndWatch(
	ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (info ExtrinsicInfo,
	err error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return info, fmt.Errorf("failed to get DID: %w", err)
	}

	txHash, bn, sig, err := a.SubmitExtrinsic(ctx, meta, c, krp)
	if err != nil {
		return info, fmt.Errorf("failed to submit extrinsic: %w", err)
	}

	s, err := getSignature(sig)
	if err != nil {
		return info, fmt.Errorf("failed to get signature: %w", err)
	}

	task := fmt.Sprintf("cent_chain_tx_status-%s", txHash.Hex())
	a.dispatcher.RegisterRunnerFunc(task, func([]interface{}, map[string]interface{}) (result interface{}, err error) {
		bh, err := a.sapi.GetBlockHash(uint64(bn))
		if err != nil {
			return nil, fmt.Errorf("failed to get block hash: %w", err)
		}

		block, err := a.sapi.GetBlock(bh)
		if err != nil {
			return nil, fmt.Errorf("failed to get block: %w", err)
		}

		extIdx := isExtrinsicSignatureInBlock(s, block.Block)
		if extIdx < 0 {
			log.Debugf("Extrinsic %s not found in block %d, trying in next block...", txHash.Hex(), bn)
			bn++
			return nil, fmt.Errorf("extrinsic %s not found in block %d", txHash.Hex(), bn)
		}

		events, err := a.checkExtrinsicEventSuccess(bh, extIdx)
		if err != nil {
			return nil, err
		}

		info := ExtrinsicInfo{
			Hash:      txHash,
			BlockHash: bh,
			Index:     uint(extIdx),
			Events:    events,
		}
		return info, nil
	})

	job := gocelery.NewRunnerFuncJob("", task, nil, nil, time.Time{})
	res, err := a.dispatcher.Dispatch(did, job)
	if err != nil {
		return info, fmt.Errorf("failed to dispatch job: %w", err)
	}

	result, err := res.Await(context.Background())
	if err != nil {
		return info, err
	}

	return result.(ExtrinsicInfo), nil
}

const (
	ExtrinsicSuccessEventName = "System.ExtrinsicSuccess"
	ExtrinsicFailedEventName  = "System.ExtrinsicFailed"
)

func (a *api) checkExtrinsicEventSuccess(
	blockHash types.Hash,
	extrinsicIdx int,
) ([]*parser.Event, error) {
	events, err := a.eventRetriever.GetEvents(blockHash)

	if err != nil {
		return nil, fmt.Errorf("event retrieval error: %w", err)
	}

	for _, event := range events {
		switch {
		case event.Name == ExtrinsicSuccessEventName &&
			event.Phase.IsApplyExtrinsic &&
			event.Phase.AsApplyExtrinsic == uint32(extrinsicIdx):
			return events, nil
		case event.Name == ExtrinsicFailedEventName &&
			event.Phase.IsApplyExtrinsic &&
			event.Phase.AsApplyExtrinsic == uint32(extrinsicIdx):
			return nil, fmt.Errorf("extrinsic with index %d failed", extrinsicIdx)
		}
	}

	return nil, errors.New("should not have reached this step: %v", events)
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
	err = a.sapi.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return 0, err
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
	return types.Signature{}, errors.New("MultiSignature not supported")
}

func isExtrinsicSignatureInBlock(extSign types.Signature, block types.Block) int {
	found := -1
	for idx, xx := range block.Extrinsics {
		if xx.Signature.Signature.IsSr25519 && xx.Signature.Signature.AsSr25519 == extSign {
			found = idx
			break
		}

		if xx.Signature.Signature.IsEd25519 && xx.Signature.Signature.AsEd25519 == extSign {
			found = idx
			break
		}
	}
	return found
}
