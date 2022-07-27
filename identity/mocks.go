//go:build unit || integration || testworld

package identity

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type MockFactory struct {
	mock.Mock
	Factory
}

func (m *MockFactory) CreateIdentity(ethAccount string, keys []Key) (transaction *types.Transaction, err error) {
	args := m.Called(ethAccount, keys)
	txn, _ := args.Get(0).(*types.Transaction)
	return txn, args.Error(1)
}

func (m *MockFactory) IdentityExists(did DID) (exists bool, err error) {
	args := m.Called(did)
	exists, _ = args.Get(0).(bool)
	return exists, args.Error(1)
}

func (m *MockFactory) NextIdentityAddress() (DID, error) {
	args := m.Called()
	did, _ := args.Get(0).(DID)
	return did, args.Error(1)
}
