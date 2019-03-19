// +build unit

package txv1

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

type mockConfig struct{}

func (mockConfig) GetEthereumContextWaitTimeout() time.Duration {
	panic("implement me")
}

func TestService_ExecuteWithinTX_happy(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	srv := ctx[transactions.BootstrappedService].(transactions.Manager)
	tid, done, err := srv.ExecuteWithinTX(context.Background(), cid, transactions.NilTxID(), "", func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, err chan<- error) {
		err <- nil
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetTransaction(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, transactions.Success, trn.Status)
}

func TestService_ExecuteWithinTX_err(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	srv := ctx[transactions.BootstrappedService].(transactions.Manager)
	tid, done, err := srv.ExecuteWithinTX(context.Background(), cid, transactions.NilTxID(), "", func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, err chan<- error) {
		err <- errors.New("dummy")
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetTransaction(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, transactions.Failed, trn.Status)
}

func TestService_ExecuteWithinTX_ctxDone(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	srv := ctx[transactions.BootstrappedService].(transactions.Manager)
	ctx, canc := context.WithCancel(context.Background())
	tid, done, err := srv.ExecuteWithinTX(ctx, cid, transactions.NilTxID(), "", func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, err chan<- error) {
		// doing nothing
	})
	canc()
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetTransaction(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, transactions.Pending, trn.Status)
	assert.Contains(t, trn.Logs[0].Message, "stopped because of context close")
}

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[transactions.BootstrappedRepo].(transactions.Repository)
	srv := ctx[transactions.BootstrappedService].(transactions.Manager)

	cid := testingidentity.GenerateRandomDID()
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(cid[:], bytes))
	txn := transactions.NewTransaction(cid, "Some transaction")

	// no transaction
	txs, err := srv.GetTransactionStatus(cid, txn.ID)
	assert.Nil(t, txs)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(transactions.ErrTransactionMissing, err))

	txn.Status = transactions.Pending
	assert.Nil(t, repo.Save(txn))

	// pending with no log
	txs, err = srv.GetTransactionStatus(cid, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.TransactionId, txn.ID.String())
	assert.Equal(t, string(transactions.Pending), txs.Status)
	assert.Empty(t, txs.Message)
	tm, err := utils.ToTimestamp(txn.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, tm, txs.LastUpdated)

	log := transactions.NewLog("action", "some message")
	txn.Logs = append(txn.Logs, log)
	txn.Status = transactions.Success
	assert.Nil(t, repo.Save(txn))

	// log with message
	txs, err = srv.GetTransactionStatus(cid, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.TransactionId, txn.ID.String())
	assert.Equal(t, string(transactions.Success), txs.Status)
	assert.Equal(t, log.Message, txs.Message)
	tm, err = utils.ToTimestamp(log.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, tm, txs.LastUpdated)
}

func TestService_CreateTransaction(t *testing.T) {
	srv := ctx[transactions.BootstrappedService].(extendedManager)
	cid := testingidentity.GenerateRandomDID()
	tx, err := srv.createTransaction(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.DID.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[transactions.BootstrappedService].(extendedManager)
	repo := ctx[transactions.BootstrappedRepo].(transactions.Repository)
	cid := testingidentity.GenerateRandomDID()

	// failed
	tx, err := srv.createTransaction(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.DID.String())
	tx.Status = transactions.Failed
	assert.NoError(t, repo.Save(tx))
	assert.Error(t, srv.WaitForTransaction(cid, tx.ID))

	// success
	tx.Status = transactions.Success
	assert.NoError(t, repo.Save(tx))
	assert.NoError(t, srv.WaitForTransaction(cid, tx.ID))
}
