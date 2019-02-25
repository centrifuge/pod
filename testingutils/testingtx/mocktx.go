// +build integration unit

package testingtx

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/mock"
)

type MockTxManager struct {
	mock.Mock
}

func (m MockTxManager) GetDefaultTaskTimeout() time.Duration {
	panic("implement me")
}

func (m MockTxManager) UpdateTaskStatus(accountID identity.DID, id transactions.TxID, status transactions.Status, taskName, message string) error {
	panic("implement me")
}

func (m MockTxManager) ExecuteWithinTX(ctx context.Context, accountID identity.DID, existingTxID transactions.TxID, desc string, work func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, err chan<- error)) (txID transactions.TxID, done chan bool, err error) {
	args := m.Called(ctx, accountID, existingTxID, desc, work)
	return args.Get(0).(transactions.TxID), args.Get(1).(chan bool), args.Error(2)
}

func (MockTxManager) GetTransaction(accountID identity.DID, id transactions.TxID) (*transactions.Transaction, error) {
	panic("implement me")
}

func (MockTxManager) GetTransactionStatus(accountID identity.DID, id transactions.TxID) (*transactionspb.TransactionStatusResponse, error) {
	panic("implement me")
}

func (MockTxManager) WaitForTransaction(accountID identity.DID, txID transactions.TxID) error {
	panic("implement me")
}
