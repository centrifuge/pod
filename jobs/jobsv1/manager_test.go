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
	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type mockConfig struct{}

func (mockConfig) GetTaskValidDuration() time.Duration {
	panic("implement me")
}

var sendChan chan notification.Message

type mockSender struct{}

// Send mocks failure and returns to channel
func (mockSender) Send(ctx context.Context, ntf notification.Message) (notification.Status, error) {
	sendChan <- ntf
	return notification.Failure, nil
}

func TestService_ExecuteWithinTX_happy(t *testing.T) {
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	jobID, done, err := srv.ExecuteWithinJob(context.Background(), did, jobs.NilJobID(), "", func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		err <- nil
	})
	assert.NoError(t, err)
	doneErr := <-done
	assert.NoError(t, doneErr)
	assert.NotNil(t, jobID)
	job, err := srv.GetJob(did, jobID)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Success, job.Status)
}

func TestService_ExecuteWithinTX_err(t *testing.T) {
	errStr := "dummy"
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	msrv := srv.(*manager)
	mngr := NewManager(msrv.config, msrv.repo)
	omgr := mngr.(*manager)
	omgr.notifier = &mockSender{}
	sendChan = make(chan notification.Message)
	jobID, done, err := omgr.ExecuteWithinJob(context.Background(), did, jobs.NilJobID(), "SomeTask", func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		err <- errors.New(errStr)
	})
	assert.NoError(t, err)
	doneErr := <-done
	assert.Error(t, doneErr)
	ntf := <-sendChan
	assert.Equal(t, notification.JobCompleted, ntf.EventType)
	assert.Equal(t, jobs.JobDataTypeURL, ntf.DocumentType)
	assert.Equal(t, string(jobs.Failed), ntf.Status)
	assert.Equal(t, errStr, ntf.Message)
	assert.NoError(t, err)
	assert.NotNil(t, jobID)
	job, err := omgr.GetJob(did, jobID)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Failed, job.Status)
	assert.Len(t, job.Logs, 1)
	assert.Equal(t, fmt.Sprintf("%s[SomeTask]", managerLogPrefix), job.Logs[0].Action)
	assert.Equal(t, errStr, job.Logs[0].Message)
}

func TestService_ExecuteWithinTX_ctxDone(t *testing.T) {
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	ctx, canc := context.WithCancel(context.Background())
	tid, done, err := srv.ExecuteWithinJob(ctx, did, jobs.NilJobID(), "", func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		// doing nothing
	})
	canc()
	assert.NoError(t, err)
	doneErr := <-done
	assert.NoError(t, doneErr)
	assert.NotNil(t, tid)
	job, err := srv.GetJob(did, tid)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Pending, job.Status)
	assert.Contains(t, job.Logs[0].Message, "stopped because of context close")
}

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)

	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(did[:], bytes))
	job := jobs.NewJob(did, "Some transaction")

	// no transaction
	jobStatus, err := srv.GetJobStatus(did, job.ID)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	job.Status = jobs.Pending
	assert.Nil(t, repo.Save(job))

	// pending with no log
	jobStatus, err = srv.GetJobStatus(did, job.ID)
	assert.Nil(t, err)
	assert.NotNil(t, jobStatus)
	assert.Equal(t, jobStatus.JobID, job.ID.String())
	assert.Equal(t, string(jobs.Pending), jobStatus.Status)
	assert.Empty(t, jobStatus.Message)
	assert.Equal(t, job.CreatedAt, jobStatus.LastUpdated)

	log := jobs.NewLog("action", "some message")
	job.Logs = append(job.Logs, log)
	job.Status = jobs.Success
	assert.Nil(t, repo.Save(job))

	// log with message
	jobStatus, err = srv.GetJobStatus(did, job.ID)
	assert.Nil(t, err)
	assert.NotNil(t, jobStatus)
	assert.Equal(t, jobStatus.JobID, job.ID.String())
	assert.Equal(t, string(jobs.Success), jobStatus.Status)
	assert.Equal(t, log.Message, jobStatus.Message)
	assert.Equal(t, log.CreatedAt, jobStatus.LastUpdated)
}

func TestService_CreateTransaction(t *testing.T) {
	srv := ctx[jobs.BootstrappedService].(extendedManager)
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	job, err := srv.createJob(did, "test")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, did.String(), job.DID.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[jobs.BootstrappedService].(extendedManager)
	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))

	// failed
	job, err := srv.createJob(did, "test")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, did.String(), job.DID.String())
	job.Status = jobs.Failed
	assert.NoError(t, repo.Save(job))
	assert.Error(t, srv.WaitForJob(did, job.ID))

	// success
	job.Status = jobs.Success
	assert.NoError(t, repo.Save(job))
	assert.NoError(t, srv.WaitForJob(did, job.ID))
}
