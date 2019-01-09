package ethereum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"time"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
)

const (
	TransactionStatusTaskName string = "TransactionMintingTask"
	TransactionTxHashParam                string = "TxHashParam"
	TransactionTenantIDParam               string = "Tenant ID"
)

type transactionStatusTask struct {
	transactions.BaseTask
	timeout         time.Duration

	//task parameter
	txHash			string
	tenantID        identity.CentID
	blockHeight     uint64


}

func NewTransactionStatusTask(
	timeout time.Duration,
	txService transactions.Service,
) *transactionStatusTask {
	return &transactionStatusTask{
		timeout:               timeout,
		BaseTask:              transactions.BaseTask{TxService: txService},
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
		txHash: nftc.txHash,
		tenantID: nftc.tenantID,
		blockHeight: nftc.blockHeight,
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

	// parse blockHeight
	nftc.blockHeight, err = queue.ParseBlockHeight(kwargs)
	if err != nil {
		return err
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

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *transactionStatusTask) RunTask() (resp interface{}, err error) {

	isPending := true
	var tx *types.Transaction

	for isPending {

		fmt.Println("fuu")

	client := GetClient()
	tx, isPending, err = client.GetEthClient().TransactionByHash(context.Background(),common.HexToHash(nftc.txHash))

	if err != nil {
		return nil, err
	}
	}


	fmt.Println("I am a task to check a successful transaction")
	fmt.Println(tx)
	return nil, nil

}
