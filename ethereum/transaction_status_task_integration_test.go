// +build integration

package ethereum_test

import (
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/assert"
)

func enqueueJob(t *testing.T, txHash string) (transactions.Service, identity.CentID, *transactions.Transaction) {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txService := ctx[transactions.BootstrappedService].(transactions.Service)

	cid := identity.RandomCentID()
	tx, err := txService.CreateTransaction(cid, "Mint NFT")

	assert.Nil(t, err, "toCentID shouldn't throw an error")

	_, err = queueSrv.EnqueueJob(ethereum.TransactionStatusTaskName, map[string]interface{}{
		transactions.TxIDParam:           tx.ID.String(),
		ethereum.TransactionAccountParam: cid.String(),
		ethereum.TransactionTxHashParam:  txHash,
	})

	time.Sleep(100 * time.Millisecond)
	return txService, cid, tx

}

func TestTransactionStatusTask_successful(t *testing.T) {
	txService, cid, tx := enqueueJob(t, "0x1")

	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a transaction should be returned")
	assert.Equal(t, string(transactions.Success), string(trans.Status), "transaction should be successful")

}

func TestTransactionStatusTask_failed(t *testing.T) {
	txService, cid, tx := enqueueJob(t, "0x2")

	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a  centrifuge transaction should be  returned")
	assert.Equal(t, string(transactions.Failed), string(trans.Status), "transaction should fail")

}

func TestTransactionStatusTask_timeout_failed(t *testing.T) {
	txService, cid, tx := enqueueJob(t, "0x3")

	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a centrifuge transaction should be returned")
	assert.Equal(t, string(transactions.Pending), string(trans.Status), "transaction should be pending")

	time.Sleep(time.Second)
	trans, err = txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a centrifuge transaction should be returned")
	assert.Equal(t, string(transactions.Failed), string(trans.Status), "transaction should fail")
}
