// +build unit

package transactions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[BootstrappedRepo].(Repository)
	srv := ctx[BootstrappedService].(Service)

	identity := common.Address([20]byte{1})
	bytes := utils.RandomSlice(common.AddressLength)
	assert.Equal(t, common.AddressLength, copy(identity[:], bytes))
	txn := NewTransaction(identity, "Some transaction")

	// no transaction
	txs, err := srv.GetTransactionStatus(identity, txn.ID)
	assert.Nil(t, txs)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	txn.Status = Pending
	assert.Nil(t, repo.Save(txn))

	// pending with no log
	txs, err = srv.GetTransactionStatus(identity, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.TransactionId, txn.ID.String())
	assert.Equal(t, string(Pending), txs.Status)
	assert.Empty(t, txs.Message)
	assert.Equal(t, utils.ToTimestamp(txn.CreatedAt), txs.LastUpdated)

	log := NewLog("action", "some message")
	txn.Logs = append(txn.Logs, log)
	txn.Status = Success
	assert.Nil(t, repo.Save(txn))

	// log with message
	txs, err = srv.GetTransactionStatus(identity, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.TransactionId, txn.ID.String())
	assert.Equal(t, string(Success), txs.Status)
	assert.Equal(t, log.Message, txs.Message)
	assert.Equal(t, utils.ToTimestamp(log.CreatedAt), txs.LastUpdated)
}

func TestService_CreateTransaction(t *testing.T) {
	srv := ctx[BootstrappedService].(Service)
	identity := common.Address([20]byte{1})
	tx, err := srv.CreateTransaction(identity, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, identity.String(), tx.Identity.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[BootstrappedService].(Service)
	repo := ctx[BootstrappedRepo].(Repository)
	identity := common.Address([20]byte{1})

	// failed
	tx, err := srv.CreateTransaction(identity, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, identity.String(), tx.Identity.String())
	tx.Status = Failed
	assert.NoError(t, repo.Save(tx))
	assert.Error(t, srv.WaitForTransaction(identity, tx.ID))

	// success
	tx.Status = Success
	assert.NoError(t, repo.Save(tx))
	assert.NoError(t, srv.WaitForTransaction(identity, tx.ID))
}
