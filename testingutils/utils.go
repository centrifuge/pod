// +build integration unit

package testingutils

import (
	"github.com/centrifuge/go-centrifuge/config"
)

func MockConfigOption(key string, value interface{}) func() {
	mockedValue := config.Config().V.Get(key)
	config.Config().V.Set(key, value)
	return func() {
		config.Config().V.Set(key, mockedValue)
	}
}

type MockSubscription struct {
	ErrChan chan error
}

func (m *MockSubscription) Err() <-chan error {
	return m.ErrChan
}

func (*MockSubscription) Unsubscribe() {}
