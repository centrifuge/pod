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

func enqueueJob(t *testing.T, txHash string) (transactions.Service, identity.CentID, *transactions.Transaction, queue.TaskResult) {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txService := ctx[transactions.BootstrappedService].(transactions.Service)

	cid := identity.RandomCentID()
	tx, err := txService.CreateTransaction(cid, "Mint NFT")

	assert.Nil(t, err, "toCentID shouldn't throw an error")

	result, err := queueSrv.EnqueueJob(ethereum.TransactionStatusTaskName, map[string]interface{}{
		transactions.TxIDParam:           tx.ID.String(),
		ethereum.TransactionAccountParam: cid.String(),
		ethereum.TransactionTxHashParam:  txHash,
	})

	return txService, cid, tx, result
}

func TestTransactionStatusTask_successful(t *testing.T) {
	txService, cid, tx, result := enqueueJob(t, "0x1")

	_, err := result.Get(time.Second)
	assert.NoError(t, err)
	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a transaction should be returned")
	assert.Equal(t, string(transactions.Success), string(trans.Status), "transaction should be successful")
}

func TestTransactionStatusTask_failed(t *testing.T) {
	txService, cid, tx, result := enqueueJob(t, "0x2")

	_, err := result.Get(time.Second)
	assert.Error(t, err)
	trans, err := txService.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a  centrifuge transaction should be  returned")
	assert.Equal(t, string(transactions.Failed), string(trans.Status), "transaction should fail")
}
