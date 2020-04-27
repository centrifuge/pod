package centchain

import (
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/centrifuge/gocelery"
)

const (
	// ExtrinsicStatusTaskName contains the name of the task
	ExtrinsicStatusTaskName string = "ExtrinsicStatusTaskName"

	// TransactionExtHashParam contains the name of the parameter
	TransactionExtHashParam string = "ExtHashParam"

	// TransactionFromBlockParam contains the name of the parameter
	TransactionFromBlockParam string = "FromBlockParam"

	// TransactionExtSignatureParam contains the name of the parameter
	TransactionExtSignatureParam string = "ExtSignatureParam"

	// TransactionAccountParam contains the name  of the account
	TransactionAccountParam string = "Account ID"
)

// ExtrinsicStatusTask struct for the task to check a cent-chain transaction
type ExtrinsicStatusTask struct {
	jobsv1.BaseTask
	intervalRetry time.Duration
	maxRetries    int

	//state
	getBlockHash      func(blockNumber uint64) (types.Hash, error)
	getBlock          func(blockHash types.Hash) (*types.SignedBlock, error)
	getMetadataLatest func() (*types.Metadata, error)
	getStorage        func(key types.StorageKey, target interface{}, blockHash types.Hash) error

	//extHash is the cent-chain extrinsic hash
	extHash string
	//fromBlock is the start block to look for extrinsic
	fromBlock uint32
	//extSignature matching signature of extrinsic in block
	extSignature types.Signature
	accountID    identity.DID

	//event filter
	eventName     string
	eventValueIdx int
}

// NewExtrinsicStatusTask returns a the struct for the task
func NewExtrinsicStatusTask(
	intervalRetry time.Duration,
	maxRetries int,
	txService jobs.Manager,
	getBlockHash func(blockNumber uint64) (types.Hash, error),
	getBlock func(blockHash types.Hash) (*types.SignedBlock, error),
	getMetadataLatest func() (*types.Metadata, error),
	getStorage func(key types.StorageKey, target interface{}, blockHash types.Hash) error,
) *ExtrinsicStatusTask {
	return &ExtrinsicStatusTask{
		intervalRetry:     intervalRetry,
		maxRetries:        maxRetries,
		BaseTask:          jobsv1.BaseTask{JobManager: txService},
		getBlockHash:      getBlockHash,
		getBlock:          getBlock,
		getMetadataLatest: getMetadataLatest,
		getStorage:        getStorage,
	}
}

// TaskTypeName returns ExtrinsicStatusTaskName
func (est *ExtrinsicStatusTask) TaskTypeName() string {
	return ExtrinsicStatusTaskName
}

// Copy returns a new instance of extrinsicStatusTask
func (est *ExtrinsicStatusTask) Copy() (gocelery.CeleryTask, error) {
	return &ExtrinsicStatusTask{
		intervalRetry:     est.intervalRetry,
		maxRetries:        est.maxRetries,
		BaseTask:          jobsv1.BaseTask{JobManager: est.JobManager},
		extHash:           est.extHash,
		fromBlock:         est.fromBlock,
		extSignature:      est.extSignature,
		accountID:         est.accountID,
		getMetadataLatest: est.getMetadataLatest,
		getStorage:        est.getStorage,
		getBlock:          est.getBlock,
		getBlockHash:      est.getBlockHash,
	}, nil
}

// ParseKwargs - define a method to parse gocelery params
func (est *ExtrinsicStatusTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	err = est.ParseJobID(est.TaskTypeName(), kwargs)
	if err != nil {
		return err
	}

	accountID, ok := kwargs[TransactionAccountParam].(string)
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("missing account ID"))
	}

	est.accountID, err = identity.NewDIDFromString(accountID)
	if err != nil {
		return err
	}

	// parse extHash
	extHash, ok := kwargs[TransactionExtHashParam]
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("undefined kwarg "+TransactionExtHashParam))
	}
	est.extHash, ok = extHash.(string)
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("malformed kwarg [%s]", TransactionExtHashParam))
	}

	// parse fromBlock
	fromBlock, ok := kwargs[TransactionFromBlockParam]
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("undefined kwarg "+TransactionFromBlockParam))
	}

	floatFromBlock, ok := fromBlock.(float64)
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("malformed kwarg [%s]", TransactionFromBlockParam))
	}
	est.fromBlock = uint32(floatFromBlock)

	// parse extSignature
	extSignature, ok := kwargs[TransactionExtSignatureParam]
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("undefined kwarg "+TransactionExtSignatureParam))
	}
	sSignature, ok := extSignature.(string)
	if !ok {
		return errors.NewTypedError(ErrExtrinsic, errors.New("malformed kwarg [%s]", TransactionExtSignatureParam))
	}
	bSign, err := types.HexDecodeString(sSignature)
	if err != nil {
		return errors.NewTypedError(ErrExtrinsic, err)
	}
	est.extSignature = types.NewSignature(bSign)

	return nil
}

// RunTask calls listens to events from cent-chain client related to extrinsicStatusTask and records result.
func (est *ExtrinsicStatusTask) RunTask() (resp interface{}, err error) {
	var jobValue *jobs.JobValue
	defer func() {
		err = est.UpdateJobWithValue(est.accountID, est.TaskTypeName(), err, jobValue)
	}()

	return est.processRunTask()
}

func (est *ExtrinsicStatusTask) processRunTask() (resp interface{}, err error) {
	var current int
	for {

		if current >= est.maxRetries {
			return nil, errors.NewTypedError(ErrCentChainTransaction, errors.New("max tries reached for extrinsic %s: %v", est.extHash, err))
		}

		nhBlock, err := est.getBlockHash(uint64(est.fromBlock))
		if err != nil {
			if err.Error() == ErrBlockNotReady.Error() {
				log.Warningf("Block %d not ready yet, trying again...", est.fromBlock)
				time.Sleep(est.intervalRetry)
				current++
				continue
			}
			return nil, err
		}

		nBlock, err := est.getBlock(nhBlock)
		if err != nil {
			return nil, err
		}

		foundIdx := isExtrinsicSignatureInBlock(est.extSignature, nBlock.Block)
		if foundIdx == -1 { // Not found in block, try next block
			log.Warningf("Extrinsic %s not found in block %d, trying in next block...", est.extHash, est.fromBlock)
			est.fromBlock = est.fromBlock + 1 // Increment block number for next iteration
			current++
			time.Sleep(est.intervalRetry)
			continue
		}

		log.Infof("Found extrinsic %s in block %d", est.extHash, est.fromBlock)

		return nil, est.parseExtrinsicStatus(nhBlock, foundIdx)
	}

}

func (est *ExtrinsicStatusTask) parseExtrinsicStatus(nhBlock types.Hash, foundIdx int) error {
	meta, err := est.getMetadataLatest()
	if err != nil {
		return err
	}

	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return err
	}

	var er types.EventRecordsRaw
	err = est.getStorage(key, &er, nhBlock)
	if err != nil {
		return err
	}

	e := Events{}
	err = er.DecodeEventRecords(meta, &e)
	if err != nil {
		return err
	}

	// Check success events
	for _, es := range e.System_ExtrinsicSuccess {
		if es.Phase.IsApplyExtrinsic && es.Phase.AsApplyExtrinsic == uint32(foundIdx) {
			return nil // Success executing extrinsic
		}
	}

	// Otherwise, check failure events
	for _, es := range e.System_ExtrinsicFailed {
		if es.Phase.IsApplyExtrinsic && es.Phase.AsApplyExtrinsic == uint32(foundIdx) {
			return errors.New("extrinsic %s failed %v", est.extHash, es.DispatchError) //Failure executing extrinsic
		}
	}

	return errors.New("should not have reached this step: %v", e)
}

func isExtrinsicSignatureInBlock(extSign types.Signature, block types.Block) int {
	found := -1
	for idx, xx := range block.Extrinsics {
		if xx.Signature.Signature.AsSr25519 == extSign {
			found = idx
			break
		}
	}
	return found
}
