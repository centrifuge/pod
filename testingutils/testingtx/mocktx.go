// +build integration unit

package testingtx

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
)

type MockTxManager struct {
	mock.Mock
}

func (m MockTxManager) GetDefaultTaskTimeout() time.Duration {
	panic("implement me")
}

func (m MockTxManager) UpdateTaskStatus(accountID identity.DID, id uuid.UUID, status transactions.Status, taskName, message string) error {
	panic("implement me")
}

func (m MockTxManager) ExecuteWithinTX(ctx context.Context, accountID identity.DID, existingTxID uuid.UUID, desc string, work func(accountID identity.DID, txID uuid.UUID, txMan transactions.Manager, err chan<- error)) (txID uuid.UUID, done chan bool, err error) {
	args := m.Called(ctx, accountID, existingTxID, desc, work)
	return args.Get(0).(uuid.UUID), args.Get(1).(chan bool), args.Error(2)
}

func (MockTxManager) GetTransaction(accountID identity.DID, id uuid.UUID) (*transactions.Transaction, error) {
	panic("implement me")
}

func (MockTxManager) GetTransactionStatus(accountID identity.DID, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error) {
	panic("implement me")
}

func (MockTxManager) WaitForTransaction(accountID identity.DID, txID uuid.UUID) error {
	panic("implement me")
}
