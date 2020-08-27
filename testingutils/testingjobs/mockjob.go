// +build integration unit

package testingjobs

import (
	"context"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/stretchr/testify/mock"
)

type MockJobManager struct {
	mock.Mock
	jobs.Manager
}

func (m MockJobManager) ExecuteWithinJob(ctx context.Context, accountID identity.DID, existingTxID jobs.JobID, desc string, work func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, err chan<- error)) (txID jobs.JobID, done chan error, err error) {
	args := m.Called(ctx, accountID, existingTxID, desc, work)
	return args.Get(0).(jobs.JobID), args.Get(1).(chan error), args.Error(2)
}

func (m MockJobManager) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	args := m.Called(account, id)
	resp, _ := args.Get(0).(jobs.StatusResponse)
	return resp, args.Error(1)
}

func (m MockJobManager) UpdateTaskStatus(accountID identity.DID, id jobs.JobID, status jobs.Status, taskName, message string) error {
	args := m.Called(accountID, id, status, taskName, message)
	return args.Error(0)
}
