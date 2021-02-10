// +build integration unit testworld

package testingutils

import (
	"github.com/centrifuge/go-centrifuge/config"
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
