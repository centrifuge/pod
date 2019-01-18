// +build unit

package transactions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[BootstrappedRepo].(Repository)
	srv := ctx[BootstrappedService].(Service)

	cid := identity.RandomCentID()
	bytes := utils.RandomSlice(identity.CentIDLength)
	assert.Equal(t, identity.CentIDLength, copy(cid[:], bytes))
	txn := NewTransaction(cid, "Some transaction")

	// no transaction
	txs, err := srv.GetTransactionStatus(cid, txn.ID)
	assert.Nil(t, txs)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	txn.Status = Pending
	assert.Nil(t, repo.Save(txn))

	// pending with no log
	txs, err = srv.GetTransactionStatus(cid, txn.ID)
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
	txs, err = srv.GetTransactionStatus(cid, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.TransactionId, txn.ID.String())
	assert.Equal(t, string(Success), txs.Status)
	assert.Equal(t, log.Message, txs.Message)
	assert.Equal(t, utils.ToTimestamp(log.CreatedAt), txs.LastUpdated)
}

func TestService_CreateTransaction(t *testing.T) {
	srv := ctx[BootstrappedService].(Service)
	cid := identity.RandomCentID()
	tx, err := srv.CreateTransaction(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.CID.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[BootstrappedService].(Service)
	repo := ctx[BootstrappedRepo].(Repository)
	cid := identity.RandomCentID()

	// failed
	tx, err := srv.CreateTransaction(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.CID.String())
	tx.Status = Failed
	assert.NoError(t, repo.Save(tx))
	assert.Error(t, srv.WaitForTransaction(cid, tx.ID))

	// success
	tx.Status = Success
	assert.NoError(t, repo.Save(tx))
	assert.NoError(t, srv.WaitForTransaction(cid, tx.ID))
}

func TestService_RegisterHandler(t *testing.T) {
	srv := ctx[BootstrappedService].(Service)
	cid := identity.RandomCentID()
	tx, err := srv.CreateTransaction(cid, "")
	assert.NoError(t, err)
	var called int
	srv.RegisterHandler(tx.ID, func(status Status) error {
		assert.Equal(t, Success, status)
		called++
		return nil
	})

	tx.Logs = append(tx.Logs, NewLog("", ""))
	tx.Status = Success
	assert.NoError(t, srv.SaveTransaction(tx))
	assert.Equal(t, 1, called)
	assert.NoError(t, srv.SaveTransaction(tx))
	assert.Equal(t, 1, called)

	// errors
	srv.RegisterHandler(tx.ID, func(status Status) error {
		assert.Equal(t, Success, status)
		return errors.New("failed handler")
	})

	assert.Len(t, tx.Logs, 1)
	assert.NoError(t, srv.SaveTransaction(tx))
	tx, err = srv.GetTransaction(cid, tx.ID)
	assert.NoError(t, err)
	assert.Len(t, tx.Logs, 2)
	assert.Contains(t, tx.Logs[1].Message, "failed handler")
}
