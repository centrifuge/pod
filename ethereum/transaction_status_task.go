package ethereum

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// EthTXStatusTaskName contains the name of the task
	EthTXStatusTaskName string = "EthTXStatusTaskName"

	// TransactionTxHashParam contains the name  of the parameter
	TransactionTxHashParam string = "TxHashParam"

	// TransactionAccountParam contains the name  of the account
	TransactionAccountParam string = "Account ID"

	// TransactionEventName contains the name of the event filtered
	TransactionEventName string = "TxEventName"

	// TransactionEventValueIdx contains the index of the position of the event value
	TransactionEventValueIdx string = "TxEventValueIdx"
)

// WatchTransaction holds the transaction status received form chain event
type WatchTransaction struct {
	Status uint64
	Error  error
}

// TransactionStatusTask is struct for the task to check an Ethereum transaction
type TransactionStatusTask struct {
	jobsv1.BaseTask
	timeout time.Duration

	//state
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	transactionByHash     func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	transactionReceipt    func(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	//txHash is the id of an Ethereum transaction
	txHash    string
	accountID identity.DID

	//event filter
	eventName     string
	eventValueIdx int
}

// NewTransactionStatusTask returns a the struct for the task
func NewTransactionStatusTask(
	timeout time.Duration,
	txService jobs.Manager,
	transactionByHash func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error),
	transactionReceipt func(ctx context.Context, txHash common.Hash) (*types.Receipt, error),
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),

) *TransactionStatusTask {
	return &TransactionStatusTask{
		timeout:               timeout,
		BaseTask:              jobsv1.BaseTask{JobManager: txService},
		ethContextInitializer: ethContextInitializer,
		transactionByHash:     transactionByHash,
		transactionReceipt:    transactionReceipt,
	}
}

// TaskTypeName returns mintingConfirmationTaskName
func (tst *TransactionStatusTask) TaskTypeName() string {
	return EthTXStatusTaskName
}

// Copy returns a new instance of mintingConfirmationTask
func (tst *TransactionStatusTask) Copy() (gocelery.CeleryTask, error) {
	return &TransactionStatusTask{
		timeout:               tst.timeout,
		txHash:                tst.txHash,
		accountID:             tst.accountID,
		transactionByHash:     tst.transactionByHash,
		transactionReceipt:    tst.transactionReceipt,
		ethContextInitializer: tst.ethContextInitializer,
		BaseTask:              jobsv1.BaseTask{JobManager: tst.JobManager},
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (tst *TransactionStatusTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	err = tst.ParseJobID(tst.TaskTypeName(), kwargs)
	if err != nil {
		return err
	}

	accountID, ok := kwargs[TransactionAccountParam].(string)
	if !ok {
		return errors.NewTypedError(ErrEthTransaction, errors.New("missing account ID"))
	}

	tst.accountID, err = identity.NewDIDFromString(accountID)
	if err != nil {
		return err
	}

	// parse txHash
	txHash, ok := kwargs[TransactionTxHashParam]
	if !ok {
		return errors.NewTypedError(ErrEthTransaction, errors.New("undefined kwarg "+TransactionTxHashParam))
	}
	tst.txHash, ok = txHash.(string)
	if !ok {
		return errors.NewTypedError(ErrEthTransaction, errors.New("malformed kwarg [%s]", TransactionTxHashParam))
	}

	// parse txEventName and index
	txEventName, ok := kwargs[TransactionEventName]
	if ok {
		tst.eventName, ok = txEventName.(string)
		if !ok {
			return errors.NewTypedError(ErrEthTransaction, errors.New("malformed kwarg [%s]", TransactionEventName))
		}
		txEventValueIdx, ok := kwargs[TransactionEventValueIdx]
		if !ok {
			return errors.NewTypedError(ErrEthTransaction, errors.New("undefined kwarg "+TransactionEventValueIdx))
		}
		tst.eventValueIdx, err = GetInt(txEventValueIdx)
		if err != nil {
			return err
		}
	}

	// override TimeoutParam if provided
	tdRaw, ok := kwargs[queue.TimeoutParam]
	if ok {
		td, err := queue.GetDuration(tdRaw)
		if err != nil {
			return errors.NewTypedError(ErrEthTransaction, errors.New("malformed kwarg [%s] because [%s]", queue.TimeoutParam, err.Error()))
		}
		tst.timeout = td
	}

	return nil
}

// GetInt converts key interface (float64) to int (used queueing only)
func GetInt(key interface{}) (int, error) {
	f64, ok := key.(float64)
	if !ok {
		return 0, errors.NewTypedError(ErrEthTransaction, errors.New("Could not parse interface to float64"))
	}
	return int(f64), nil
}

// getEventsFromTransactionReceipt returns all events that are indexed
// note that events that are not indexed will not be parsed at the moment
func (tst *TransactionStatusTask) getEventValueFromTransactionReceipt(ctx context.Context, txHash string, event string, idxValue int) (value []byte, err error) {
	receipt, err := tst.transactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}
	for _, v := range receipt.Logs {
		if (len(v.Topics) > 0) && v.Topics[0].Hex() == hexutil.Encode(crypto.Keccak256([]byte(event))) {
			if idxValue < len(v.Topics) {
				return v.Topics[idxValue+1].Bytes(), nil
			}
		}
	}
	return nil, errors.NewTypedError(ErrEthTransaction, errors.New("Event [%s] with value idx [%d] not found", event, idxValue))
}

func (tst *TransactionStatusTask) isTransactionSuccessful(ctx context.Context, txHash string) error {
	receipt, err := tst.transactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return ErrTransactionFailed
	}

	return nil
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (tst *TransactionStatusTask) RunTask() (resp interface{}, err error) {
	var jobValue *jobs.JobValue
	ctx, cancelF := tst.ethContextInitializer(tst.timeout)
	defer cancelF()
	defer func() {
		err = tst.UpdateJobWithValue(tst.accountID, tst.TaskTypeName(), err, jobValue)
	}()

	_, isPending, err := tst.transactionByHash(ctx, common.HexToHash(tst.txHash))
	if err != nil {
		// if the tx is not propagated, this will error out with "Not found"
		// lets retry in this scenario as well
		if err == ethereum.NotFound {
			err = gocelery.ErrTaskRetryable
		}
		return nil, err
	}

	if isPending {
		return nil, gocelery.ErrTaskRetryable
	}

	err = tst.isTransactionSuccessful(ctx, tst.txHash)
	if err != nil {
		if err != ErrTransactionFailed {
			err = gocelery.ErrTaskRetryable
		}
		return nil, err
	}

	if tst.eventName != "" {
		v, err := tst.getEventValueFromTransactionReceipt(ctx, tst.txHash, tst.eventName, tst.eventValueIdx)
		if err != nil {
			return nil, err
		}
		log.Infof("Value [%x] found for Event [%s]\n", v, tst.eventName)
		jobValue = &jobs.JobValue{Key: tst.eventName, Value: v}
	}

	return nil, nil
}
