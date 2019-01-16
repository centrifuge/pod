package ethereum

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// TransactionStatusTaskName contains the name of the task
	TransactionStatusTaskName string = "TransactionStatusTask"
	// TransactionTxHashParam contains the name  of the parameter
	TransactionTxHashParam string = "TxHashParam"
	// TransactionAccountParam contains the name  of the account
	TransactionAccountParam  string = "Account ID"
	transactionStatusSuccess uint64 = 1

	// ErrTransactionFailed error when transaction failed
	ErrTransactionFailed = errors.Error("Transaction failed")
)

// TransactionStatusTask is struct for the task to check an Ethereum transaction
type TransactionStatusTask struct {
	transactions.BaseTask
	timeout time.Duration
	//state
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	transactionByHash     func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	transactionReceipt    func(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	//txHash is the id of an Ethereum transaction
	txHash   string
	tenantID identity.CentID
}

// NewTransactionStatusTask returns a the struct for the task
func NewTransactionStatusTask(
	timeout time.Duration,
	txService transactions.Service,
	transactionByHash func(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error),
	transactionReceipt func(ctx context.Context, txHash common.Hash) (*types.Receipt, error),
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),

) *TransactionStatusTask {
	return &TransactionStatusTask{
		timeout:               timeout,
		BaseTask:              transactions.BaseTask{TxService: txService},
		ethContextInitializer: ethContextInitializer,
		transactionByHash:     transactionByHash,
		transactionReceipt:    transactionReceipt,
	}
}

// TaskTypeName returns mintingConfirmationTaskName
func (nftc *TransactionStatusTask) TaskTypeName() string {
	return TransactionStatusTaskName
}

// Copy returns a new instance of mintingConfirmationTask
func (nftc *TransactionStatusTask) Copy() (gocelery.CeleryTask, error) {
	return &TransactionStatusTask{
		timeout:               nftc.timeout,
		txHash:                nftc.txHash,
		tenantID:              nftc.tenantID,
		transactionByHash:     nftc.transactionByHash,
		transactionReceipt:    nftc.transactionReceipt,
		ethContextInitializer: nftc.ethContextInitializer,
		BaseTask:              transactions.BaseTask{TxService: nftc.TxService},
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *TransactionStatusTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	err = nftc.ParseTransactionID(kwargs)
	if err != nil {
		return err
	}

	tenantID, ok := kwargs[TransactionAccountParam].(string)
	if !ok {
		return errors.New("missing tenant ID")
	}

	nftc.tenantID, err = identity.CentIDFromString(tenantID)
	if err != nil {
		return err
	}

	// parse txHash
	txHash, ok := kwargs[TransactionTxHashParam]
	if !ok {
		return errors.New("undefined kwarg " + TransactionTxHashParam)
	}
	nftc.txHash, ok = txHash.(string)
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
		nftc.timeout = td
	}

	return nil
}

func (nftc *TransactionStatusTask) isTransactionSuccessful(ctx context.Context, txHash string) error {
	receipt, err := nftc.transactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	if receipt.Status != transactionStatusSuccess {
		return ErrTransactionFailed
	}

	return nil
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *TransactionStatusTask) RunTask() (resp interface{}, err error) {
	ctx, cancelF := nftc.ethContextInitializer(nftc.timeout)
	defer cancelF()
	defer func() {
		if err != nil {
			log.Infof("Transaction failed: %v\n", nftc.txHash)
		} else {
			log.Infof("Transaction successful:%v\n", nftc.txHash)
		}

		err = nftc.UpdateTransaction(nftc.tenantID, nftc.TaskTypeName(), err)
	}()

	_, isPending, err := nftc.transactionByHash(ctx, common.HexToHash(nftc.txHash))
	if err != nil {
		return nil, err
	}

	if isPending {
		return nil, gocelery.ErrTaskRetryable
	}

	err = nftc.isTransactionSuccessful(ctx, nftc.txHash)
	if err == nil {
		return nil, nil
	}

	if err != ErrTransactionFailed {
		return nil, gocelery.ErrTaskRetryable
	}

	return nil, err
}
