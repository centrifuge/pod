package ethereum

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
)

const (
	TransactionStatusTaskName string = "TransactionMintingTask"
	TransactionTxHashParam    string = "TxHashParam"
	TransactionTenantIDParam  string = "Tenant ID"
	TransactionStatusSuccess  uint64 = 1
)

type transactionStatusTask struct {
	transactions.BaseTask
	timeout time.Duration
	//state
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)

	//task parameter
	txHash   string
	tenantID identity.CentID
}

func NewTransactionStatusTask(
	timeout time.Duration,
	txService transactions.Service,
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),

) *transactionStatusTask {
	return &transactionStatusTask{
		timeout:               timeout,
		BaseTask:              transactions.BaseTask{TxService: txService},
		ethContextInitializer: ethContextInitializer,
	}
}

// TaskTypeName returns mintingConfirmationTaskName
func (nftc *transactionStatusTask) TaskTypeName() string {
	return TransactionStatusTaskName
}

// Copy returns a new instance of mintingConfirmationTask
func (nftc *transactionStatusTask) Copy() (gocelery.CeleryTask, error) {
	return &transactionStatusTask{
		timeout:               nftc.timeout,
		txHash:                nftc.txHash,
		tenantID:              nftc.tenantID,
		ethContextInitializer: nftc.ethContextInitializer,
		BaseTask:              transactions.BaseTask{TxService: nftc.TxService},
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *transactionStatusTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	err = nftc.ParseTransactionID(kwargs)
	if err != nil {
		return err
	}

	tenantID, ok := kwargs[TransactionTenantIDParam].(string)
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

func getTransactionStatus(ctx context.Context, txHash string) (bool, error) {
	client := GetClient()
	receipt, err := client.GetEthClient().TransactionReceipt(ctx, common.HexToHash(txHash))

	if err != nil {
		return false, err
	}

	if receipt.Status == TransactionStatusSuccess {
		return true, nil
	}

	return false, errors.New("Transaction failed")

}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *transactionStatusTask) RunTask() (resp interface{}, err error) {
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

	isPending := true
	client := GetClient()
	for isPending {
		_, isPending, err = client.GetEthClient().TransactionByHash(ctx, common.HexToHash(nftc.txHash))
		if err != nil {
			return nil, err
		}

		if isPending == false {
			successful, err := getTransactionStatus(ctx, nftc.txHash)
			if err != nil {
				return nil, err
			}

			if successful {
				return nil, nil

			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, nil

}
