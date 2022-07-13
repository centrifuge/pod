//go:build unit || integration || testworld
// +build unit integration testworld

package proxy

import (
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
	Service
}

func (m *MockService) GetProxy(address string) (*Definition, error) {
	args := m.Called(address)
	acc, _ := args.Get(0).(*Definition)
	return acc, args.Error(1)
}

func (m *MockService) ProxyHasProxyType(proxyDef *Definition, proxied []byte, proxyType string) bool {
	return service{}.ProxyHasProxyType(proxyDef, proxied, proxyType)
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}
