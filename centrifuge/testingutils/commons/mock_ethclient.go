// +build integration unit

package testingcommons

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/mock"
)

type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) GetClient() *ethclient.Client {
	args := m.Called()
	c, _ := args.Get(0).(*ethclient.Client)
	return c
}

func (m *MockEthClient) GetRpcClient() *rpc.Client {
	args := m.Called()
	return args.Get(0).(*rpc.Client)
}

func (m *MockEthClient) GetHost() *url.URL {
	args := m.Called()
	return args.Get(0).(*url.URL)
}

func (m *MockEthClient) GetTxOpts(accountName string) (*bind.TransactOpts, error) {
	args := m.Called(accountName)
	return args.Get(0).(*bind.TransactOpts), args.Error(1)
}

func (m *MockEthClient) SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error) {
	args := m.Called(contractMethod, opts, params)
	return args.Get(0).(*types.Transaction), args.Error(1)
}
