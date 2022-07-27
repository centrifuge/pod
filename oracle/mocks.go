//go:build integration || unit || testworld

package oracle

import "github.com/stretchr/testify/mock"

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

type MockService struct {
	mock.Mock
	Service
}
