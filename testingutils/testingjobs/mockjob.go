// +build integration unit

package testingjobs

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/jobs"
	"github.com/stretchr/testify/mock"
)

type MockJobManager struct {
	mock.Mock
}

func (m MockJobManager) GetDefaultTaskTimeout() time.Duration {
	panic("implement me")
}

func (m MockJobManager) UpdateJobWithValue(accountID identity.DID, id jobs.JobID, key string, value []byte) error {
	panic("implement me")
}

func (m MockJobManager) UpdateTaskStatus(accountID identity.DID, id jobs.JobID, status jobs.Status, taskName, message string) error {
	panic("implement me")
}

func (m MockJobManager) ExecuteWithinJob(ctx context.Context, accountID identity.DID, existingTxID jobs.JobID, desc string, work func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error)) (txID jobs.JobID, done chan bool, err error) {
	args := m.Called(ctx, accountID, existingTxID, desc, work)
	return args.Get(0).(jobs.JobID), args.Get(1).(chan bool), args.Error(2)
}

func (MockJobManager) GetJob(accountID identity.DID, id jobs.JobID) (*jobs.Job, error) {
	panic("implement me")
}

func (MockJobManager) GetJobStatus(accountID identity.DID, id jobs.JobID) (*jobspb.JobStatusResponse, error) {
	panic("implement me")
}

func (MockJobManager) WaitForJob(accountID identity.DID, txID jobs.JobID) error {
	panic("implement me")
}
