// +build integration unit

package ethereum

import (
	"context"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/mock"
)

type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) GetGethCallOpts(pending bool) (*bind.CallOpts, context.CancelFunc) {
	args := m.Called()
	c, _ := args.Get(0).(*bind.CallOpts)
	return c, func() {}
}

func (m *MockEthClient) GetEthClient() EthClient {
	args := m.Called()
	c, _ := args.Get(0).(*ethclient.Client)
	return c
}

func (m *MockEthClient) GetNodeURL() *url.URL {
	args := m.Called()
	return args.Get(0).(*url.URL)
}

func (m *MockEthClient) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockEthClient) GetTxOpts(ctx context.Context, accountName string) (*bind.TransactOpts, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*bind.TransactOpts), args.Error(1)
}

func (m *MockEthClient) SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error) {
	args := m.Called(contractMethod, opts, params)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockEthClient) SubmitTransaction(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error) {
	args := m.Called(contractMethod, opts, params)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockEthClient) TransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Transaction), args.Get(1).(bool), args.Error(2)
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}
