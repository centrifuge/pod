package ethereum

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/transactions/txv1"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	// EthTXStatusTaskName contains the name of the task
	EthTXStatusTaskName string = "EthTXStatusTaskName"

	// TransactionTxHashParam contains the name  of the parameter
	TransactionTxHashParam string = "TxHashParam"

	// TransactionAccountParam contains the name  of the account
	TransactionAccountParam string = "Account ID"
	// TransactionStatusSuccess contains the flag for a successful receipt.status
	TransactionStatusSuccess uint64 = 1

	// ErrTransactionFailed error when transaction failed
	ErrTransactionFailed = errors.Error("Transaction failed")
)

// WatchTransaction holds the transaction status received form chain event
type WatchTransaction struct {
	Status uint64
	txHash string
	Error  error
}

// TransactionStatusTask is struct for the task to check an Ethereum transaction
type TransactionStatusTask struct {
	txv1.BaseTask
	timeout time.Duration

	//state
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	transactionByHash     func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	transactionReceipt    func(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	//txHash is the id of an Ethereum transaction
	txHash    string
	accountID identity.CentID
}

// NewTransactionStatusTask returns a the struct for the task
func NewTransactionStatusTask(
	timeout time.Duration,
	txService transactions.Manager,
	transactionByHash func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error),
	transactionReceipt func(ctx context.Context, txHash common.Hash) (*types.Receipt, error),
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),

) *TransactionStatusTask {
	return &TransactionStatusTask{
		timeout:               timeout,
		BaseTask:              txv1.BaseTask{TxManager: txService},
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
		BaseTask:              txv1.BaseTask{TxManager: tst.TxManager},
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (tst *TransactionStatusTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	err = tst.ParseTransactionID(tst.TaskTypeName(), kwargs)
	if err != nil {
		return err
	}

	accountID, ok := kwargs[TransactionAccountParam].(string)
	if !ok {
		return errors.New("missing account ID")
	}

	tst.accountID, err = identity.CentIDFromString(accountID)
	if err != nil {
		return err
	}

	// parse txHash
	txHash, ok := kwargs[TransactionTxHashParam]
	if !ok {
		return errors.New("undefined kwarg " + TransactionTxHashParam)
	}
	tst.txHash, ok = txHash.(string)
	if !ok {
		return errors.New("malformed kwarg [%s]", TransactionTxHashParam)
	}

	// override TimeoutParam if provided
	tdRaw, ok := kwargs[queue.TimeoutParam]
	if ok {
		td, err := queue.GetDuration(tdRaw)
		if err != nil {
			return errors.New("malformed kwarg [%s] because [%s]", queue.TimeoutParam, err.Error())
		}
		tst.timeout = td
	}

	return nil
}

func (tst *TransactionStatusTask) isTransactionSuccessful(ctx context.Context, txHash string) error {
	receipt, err := tst.transactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	if receipt.Status != TransactionStatusSuccess {
		return ErrTransactionFailed
	}

	return nil
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (tst *TransactionStatusTask) RunTask() (resp interface{}, err error) {
	ctx, cancelF := tst.ethContextInitializer(tst.timeout)
	defer cancelF()
	defer func() {
		err = tst.UpdateTransaction(tst.accountID, tst.TaskTypeName(), err)
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
	if err == nil {
		return nil, nil
	}

	if err != ErrTransactionFailed {
		return nil, gocelery.ErrTaskRetryable
	}

	return nil, err
}
