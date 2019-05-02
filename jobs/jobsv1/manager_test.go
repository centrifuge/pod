// +build unit

package jobsv1

import (
	"context"
	"fmt"
	"testing"
	"time"

	notificationpb "github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/go-centrifuge/notification"

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

var sendChan chan *notificationpb.NotificationMessage

type mockSender struct{}

// Send mocks failure and returns to channel
func (mockSender) Send(ctx context.Context, ntf *notificationpb.NotificationMessage) (notification.Status, error) {
	sendChan <- ntf
	return notification.Failure, nil
}

func TestService_ExecuteWithinTX_happy(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	jobID, done, err := srv.ExecuteWithinJob(context.Background(), did, jobs.NilJobID(), "", func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		err <- nil
	})
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, jobID)
	job, err := srv.GetJob(did, jobID)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Success, job.Status)
}

func TestService_ExecuteWithinTX_err(t *testing.T) {
	errStr := "dummy"
	did := testingidentity.GenerateRandomDID()
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	mngr := srv.(*manager)
	mngr.notifier = &mockSender{}
	sendChan = make(chan *notificationpb.NotificationMessage)
	jobID, done, err := mngr.ExecuteWithinJob(context.Background(), did, jobs.NilJobID(), "SomeTask", func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		err <- errors.New(errStr)
	})
	<-done
	ntf := <-sendChan
	assert.Equal(t, uint32(notification.JobCompleted), ntf.EventType)
	assert.Equal(t, jobs.JobDataTypeURL, ntf.DocumentType)
	assert.Equal(t, string(jobs.Failed), ntf.Status)
	assert.Equal(t, errStr, ntf.Message)
	assert.NoError(t, err)
	assert.NotNil(t, jobID)
	job, err := mngr.GetJob(did, jobID)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Failed, job.Status)
	assert.Len(t, job.Logs, 1)
	assert.Equal(t, fmt.Sprintf("%s[SomeTask]", managerLogPrefix), job.Logs[0].Action)
	assert.Equal(t, errStr, job.Logs[0].Message)
}

func TestService_ExecuteWithinTX_ctxDone(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)
	ctx, canc := context.WithCancel(context.Background())
	tid, done, err := srv.ExecuteWithinJob(ctx, did, jobs.NilJobID(), "", func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error) {
		// doing nothing
	})
	canc()
	<-done
	assert.NoError(t, err)
	assert.NotNil(t, tid)
	job, err := srv.GetJob(did, tid)
	assert.NoError(t, err)
	assert.Equal(t, jobs.Pending, job.Status)
	assert.Contains(t, job.Logs[0].Message, "stopped because of context close")
}

func TestService_GetTransaction(t *testing.T) {
	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	srv := ctx[jobs.BootstrappedService].(jobs.Manager)

	did := testingidentity.GenerateRandomDID()
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(did[:], bytes))
	job := jobs.NewJob(did, "Some transaction")

	// no transaction
	jobStatus, err := srv.GetJobStatus(did, job.ID)
	assert.Nil(t, jobStatus)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	job.Status = jobs.Pending
	assert.Nil(t, repo.Save(job))

	// pending with no log
	jobStatus, err = srv.GetJobStatus(did, job.ID)
	assert.Nil(t, err)
	assert.NotNil(t, jobStatus)
	assert.Equal(t, jobStatus.JobId, job.ID.String())
	assert.Equal(t, string(jobs.Pending), jobStatus.Status)
	assert.Empty(t, jobStatus.Message)
	tm, err := utils.ToTimestamp(job.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, tm, jobStatus.LastUpdated)

	log := jobs.NewLog("action", "some message")
	job.Logs = append(job.Logs, log)
	job.Status = jobs.Success
	assert.Nil(t, repo.Save(job))

	// log with message
	jobStatus, err = srv.GetJobStatus(did, job.ID)
	assert.Nil(t, err)
	assert.NotNil(t, jobStatus)
	assert.Equal(t, jobStatus.JobId, job.ID.String())
	assert.Equal(t, string(jobs.Success), jobStatus.Status)
	assert.Equal(t, log.Message, jobStatus.Message)
	tm, err = utils.ToTimestamp(log.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, tm, jobStatus.LastUpdated)
}

func TestService_CreateTransaction(t *testing.T) {
	srv := ctx[jobs.BootstrappedService].(extendedManager)
	did := testingidentity.GenerateRandomDID()
	job, err := srv.createJob(did, "test")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, did.String(), job.DID.String())
}

func TestService_WaitForTransaction(t *testing.T) {
	srv := ctx[jobs.BootstrappedService].(extendedManager)
	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	did := testingidentity.GenerateRandomDID()

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
