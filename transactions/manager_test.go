// +build unit

package transactions

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_ExecuteWithinTX_happy(t *testing.T) {
	cid := identity.RandomCentID()
	srv := ctx[BootstrappedService].(Manager)
	tid, done, err := srv.ExecuteWithinTX(context.Background(), cid, uuid.Nil, "", func(accountID identity.CentID, txID uuid.UUID, txMan Manager, err chan<- error) {
		err <- nil
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetTransaction(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, Success, trn.Status)
}

func TestService_ExecuteWithinTX_err(t *testing.T) {
	cid := identity.RandomCentID()
	srv := ctx[BootstrappedService].(Manager)
	tid, done, err := srv.ExecuteWithinTX(context.Background(), cid, uuid.Nil, "", func(accountID identity.CentID, txID uuid.UUID, txMan Manager, err chan<- error) {
		err <- errors.New("dummy")
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetTransaction(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, Failed, trn.Status)
}

func TestService_ExecuteWithinTX_ctxDone(t *testing.T) {
	cid := identity.RandomCentID()
	srv := ctx[BootstrappedService].(Manager)
	ctx, canc := context.WithCancel(context.Background())
	tid, done, err := srv.ExecuteWithinTX(ctx, cid, uuid.Nil, "", func(accountID identity.CentID, txID uuid.UUID, txMan Manager, err chan<- error) {
		// doing nothing
	})
	canc()
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetTransaction(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, Pending, trn.Status)
	assert.Contains(t, trn.Logs[0].Message, "stopped because of context close")
}

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[BootstrappedRepo].(Repository)
	srv := ctx[BootstrappedService].(Manager)

	cid := identity.RandomCentID()
	bytes := utils.RandomSlice(identity.CentIDLength)
	assert.Equal(t, identity.CentIDLength, copy(cid[:], bytes))
	txn := newTransaction(cid, "Some transaction")

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
	srv := ctx[BootstrappedService].(Manager)
	cid := identity.RandomCentID()
	tx, err := srv.CreateTransaction(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.CID.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[BootstrappedService].(Manager)
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
