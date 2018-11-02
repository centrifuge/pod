// +build integration unit

package testingutils

import (
	"github.com/centrifuge/go-centrifuge/config"
)

func MockConfigOption(key string, value interface{}) func() {
	mockedValue := config.Config().Get(key)
	config.Config().Set(key, value)
	return func() {
		config.Config().Set(key, mockedValue)
	}
}

type MockSubscription struct {
	ErrChan chan error
}

func (m *MockSubscription) Err() <-chan error {
	return m.ErrChan
}

func (*MockSubscription) Unsubscribe() {}
