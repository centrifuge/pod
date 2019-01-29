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

func enqueueJob(t *testing.T, txHash string) (transactions.Manager, identity.CentID, *transactions.Transaction, queue.TaskResult) {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txManager := ctx[transactions.BootstrappedService].(transactions.Manager)

	cid := identity.RandomCentID()
	tx, err := txManager.CreateTransaction(cid, "Mint NFT")

	assert.Nil(t, err, "toCentID shouldn't throw an error")

	result, err := queueSrv.EnqueueJob(ethereum.EthTXStatusTaskName, map[string]interface{}{
		transactions.TxIDParam:           tx.ID.String(),
		ethereum.TransactionAccountParam: cid.String(),
		ethereum.TransactionTxHashParam:  txHash,
	})

	return txManager, cid, tx, result
}

func TestTransactionStatusTask_successful(t *testing.T) {
	txManager, cid, tx, result := enqueueJob(t, "0x1")

	_, err := result.Get(time.Second)
	assert.NoError(t, err)
	trans, err := txManager.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a transaction should be returned")
	assert.Equal(t, string(transactions.Success), string(trans.Status), "transaction should be successful")
}

func TestTransactionStatusTask_failed(t *testing.T) {
	txManager, cid, tx, result := enqueueJob(t, "0x2")

	_, err := result.Get(time.Second)
	assert.Error(t, err)
	trans, err := txManager.GetTransaction(cid, tx.ID)
	assert.Nil(t, err, "a  centrifuge transaction should be  returned")
	assert.Equal(t, string(transactions.Failed), string(trans.Status), "transaction should fail")
}
