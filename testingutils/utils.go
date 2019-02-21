// +build integration unit

package testingutils

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

func MockConfigOption(cfg config.Configuration, key string, value interface{}) func() {
	mockedValue := cfg.Get(key)
	cfg.Set(key, value)
	return func() {
		cfg.Set(key, mockedValue)
	}
}

type MockSubscription struct {
	ErrChan chan error
}

func (m *MockSubscription) Err() <-chan error {
	return m.ErrChan
}

func (*MockSubscription) Unsubscribe() {}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) EnqueueJob(taskTypeName string, params map[string]interface{}) (queue.TaskResult, error) {
	args := m.Called(taskTypeName, params)
	res, _ := args.Get(0).(queue.TaskResult)
	return res, args.Error(1)
}

func (m *MockQueue) EnqueueJobWithMaxTries(taskTypeName string, params map[string]interface{}) (queue.TaskResult, error) {
	args := m.Called(taskTypeName, params)
	res, _ := args.Get(0).(queue.TaskResult)
	return res, args.Error(1)
}

// RandomDID return a random did for testing.
func RandomDID() identity.DID {
	randomBytes := utils.RandomSlice(common.AddressLength)
	return identity.NewDIDFromBytes(randomBytes)
}
