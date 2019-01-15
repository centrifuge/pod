// +build integration

package ethereum_test

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

func registerTransactionTask(mockClient *testingcommons.MockEthClient, timeout time.Duration) {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txService := ctx[transactions.BootstrappedService].(transactions.Service)

	ethTransTask := ethereum.NewTransactionStatusTask(timeout, txService, mockClient, ethereum.DefaultWaitForTransactionMiningContext)
	queueSrv.RegisterTaskType(ethereum.TransactionStatusTaskName, ethTransTask)

	bootstrapQueueStart()
}

func enqueueJob(t *testing.T) (transactions.Service, identity.CentID, *transactions.Transaction) {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txService := ctx[transactions.BootstrappedService].(transactions.Service)

	cid := identity.RandomCentID()
	tx, err := txService.CreateTransaction(cid, "Mint NFT")

	assert.Nil(t, err, "toCentID shouldn't throw an error")

	_, err = queueSrv.EnqueueJob(ethereum.TransactionStatusTaskName, map[string]interface{}{
		transactions.TxIDParam:           tx.ID.String(),
		ethereum.TransactionAccountParam: cid.String(),
		ethereum.TransactionTxHashParam:  "0x123",
	})

	time.Sleep(100 * time.Millisecond)
	return txService, cid, tx

}

func TestTransactionStatusTask_successful(t *testing.T) {
	mockClient := &testingcommons.MockEthClient{}
	mockClient.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil).Once()

	// transaction status = 1 (success)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{Status: 1}, nil).Once()

	registerTransactionTask(mockClient, cfg.GetEthereumContextWaitTimeout())

	txService, cid, tx := enqueueJob(t)

	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a transaction should be returned")
	assert.Equal(t, string(transactions.Success), string(trans.Status), "transaction should be successful")

}

func TestTransactionStatusTask_failed(t *testing.T) {
	mockClient := &testingcommons.MockEthClient{}
	mockClient.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, false, nil).Once()

	// transaction status = 0 (failed)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{Status: 0}, nil).Once()

	registerTransactionTask(mockClient, cfg.GetEthereumContextWaitTimeout())

	txService, cid, tx := enqueueJob(t)

	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a  centrifuge transaction should be  returned")
	assert.Equal(t, string(transactions.Failed), string(trans.Status), "transaction should fail")

}

func TestTransactionStatusTask_timeout_failed(t *testing.T) {
	mockClient := &testingcommons.MockEthClient{}

	// transactions pending too long
	mockClient.On("TransactionByHash", mock.Anything, mock.Anything).Return(&types.Transaction{}, true, nil).Once()

	registerTransactionTask(mockClient, 100*time.Millisecond)

	txService, cid, tx := enqueueJob(t)

	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a centrifuge transaction should be  returned")
	assert.Equal(t, string(transactions.Pending), string(trans.Status), "transaction should fail")

}
