//go:build unit || integration

package centchain

import (
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/client"
	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
	client.Client
}

func (c *ClientMock) Call(result interface{}, method string, args ...interface{}) error {
	arg := c.Called(result, method, args)
	res := arg.Get(0).(string)
	eres := result.(*string)
	*eres = res
	return arg.Error(1)
}

func NewClientMock(t testing.TB) *ClientMock {
	mock := &ClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
