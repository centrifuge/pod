// +build unit

package jobsv1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

type mockConfig struct{}

func (mockConfig) GetEthereumContextWaitTimeout() time.Duration {
	panic("implement me")
}

func TestService_ExecuteWithinTX_happy(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	tid, done, err := srv.ExecuteWithinJob(context.Background(), cid, jobs.NilJobID(), "", func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		err <- nil
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetJob(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Success, trn.Status)
}

func TestService_ExecuteWithinTX_err(t *testing.T) {
	errStr := "dummy"
	cid := testingidentity.GenerateRandomDID()
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	tid, done, err := srv.ExecuteWithinJob(context.Background(), cid, jobs.NilJobID(), "SomeTask", func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		err <- errors.New(errStr)
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetJob(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Failed, trn.Status)
	assert.Len(t, trn.Logs, 1)
	assert.Equal(t, fmt.Sprintf("%s[SomeTask]", managerLogPrefix), trn.Logs[0].Action)
	assert.Equal(t, errStr, trn.Logs[0].Message)
}

func TestService_ExecuteWithinTX_ctxDone(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	ctx, canc := context.WithCancel(context.Background())
	tid, done, err := srv.ExecuteWithinJob(ctx, cid, jobs.NilJobID(), "", func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		// doing nothing
	})
	canc()
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	trn, err := srv.GetJob(cid, tid)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Pending, trn.Status)
	assert.Contains(t, trn.Logs[0].Message, "stopped because of context close")
}

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)

	cid := testingidentity.GenerateRandomDID()
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(cid[:], bytes))
	txn := jobs.NewJob(cid, "Some transaction")

	// no transaction
	txs, err := srv.GetJobStatus(cid, txn.ID)
	assert.Nil(t, txs)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	txn.Status = jobs.Pending
	assert.Nil(t, repo.Save(txn))

	// pending with no log
	txs, err = srv.GetJobStatus(cid, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.JobId, txn.ID.String())
	assert.Equal(t, string(jobs.Pending), txs.Status)
	assert.Empty(t, txs.Message)
	tm, err := utils.ToTimestamp(txn.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, tm, txs.LastUpdated)

	log := jobs.NewLog("action", "some message")
	txn.Logs = append(txn.Logs, log)
	txn.Status = jobs.Success
	assert.Nil(t, repo.Save(txn))

	// log with message
	txs, err = srv.GetJobStatus(cid, txn.ID)
	assert.Nil(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, txs.JobId, txn.ID.String())
	assert.Equal(t, string(jobs.Success), txs.Status)
	assert.Equal(t, log.Message, txs.Message)
	tm, err = utils.ToTimestamp(log.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, tm, txs.LastUpdated)
}

func TestService_CreateTransaction(t *testing.T) {
	srv := ctx[jobs.BootstrappedService].(extendedManager)
	cid := testingidentity.GenerateRandomDID()
	tx, err := srv.createJob(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.DID.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[jobs.BootstrappedService].(extendedManager)
	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	cid := testingidentity.GenerateRandomDID()

	// failed
	tx, err := srv.createJob(cid, "test")
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid.String(), tx.DID.String())
	tx.Status = jobs.Failed
	assert.NoError(t, repo.Save(tx))
	assert.Error(t, srv.WaitForJob(cid, tx.ID))

	// success
	tx.Status = jobs.Success
	assert.NoError(t, repo.Save(tx))
	assert.NoError(t, srv.WaitForJob(cid, tx.ID))
}
