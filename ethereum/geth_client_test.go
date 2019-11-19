// +build unit

package ethereum

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("ethereum.txPoolAccessEnabled", false)
	cfg.Set("ethereum.intervalRetry", time.Millisecond*100)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

type MockEthCl struct {
	EthClient
	mock.Mock
}

func (m *MockEthCl) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

type MockTransactionRequest struct {
	count int
}

func (transactionRequest *MockTransactionRequest) RegisterTransaction(opts *bind.TransactOpts, transactionName string, anotherVar string) (tx *types.Transaction, err error) {
	transactionRequest.count++
	if transactionName == "otherError" {
		err = errors.New("Some other error")
	} else if transactionName == "optimisticLockingTimeout" {
		err = ErrIncNonce
	} else if transactionName == "optimisticLockingEventualSuccess" {
		if transactionRequest.count < 3 {
			err = ErrIncNonce
		}
	}

	return types.NewTransaction(1, common.Address{}, nil, 0, nil, nil), err
}

func TestInitTransactionWithRetries(t *testing.T) {
	opts := &bind.TransactOpts{From: common.HexToAddress("0x45B9c4798999FFa52e1ff1eFce9d3e45819E4158")}
	mockRequest := &MockTransactionRequest{}

	// noncer success
	mockClient := &MockEthCl{}
	mockClient.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(1), nil)
	gc := &gethClient{
		txMu:   sync.Mutex{},
		config: cfg,
		client: mockClient,
	}

	SetClient(gc)

	// Success at first
	tx, err := gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, opts, "var1", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 1, tx.Nonce(), "Nonce should equal to the one provided")
	assert.EqualValues(t, 1, mockRequest.count, "Transaction Run flag should be true")

	// Failure with non-locking error
	tx, err = gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, opts, "otherError", "var2")
	assert.EqualError(t, err, "Some other error", "Should error out")

	mockRetries := testingutils.MockConfigOption(cfg, "ethereum.maxRetries", 10)
	defer mockRetries()

	mockRequest.count = 0
	// Failure and timeout with locking error
	tx, err = gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, opts, "optimisticLockingTimeout", "var2")
	assert.Contains(t, err.Error(), ErrIncNonce, "Should error out")
	assert.EqualValues(t, 10, mockRequest.count, "Retries should be equal")

	mockRequest.count = 0
	// Success after locking race condition overcome
	tx, err = gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, opts, "optimisticLockingEventualSuccess", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 3, mockRequest.count, "Retries should be equal")
}

func TestGetGethCallOpts(t *testing.T) {
	opts, cancel := GetClient().GetGethCallOpts(true)
	assert.NotNil(t, opts)
	assert.True(t, opts.Pending)
	assert.NotNil(t, cancel)
	cancel()
}

type mockNoncer struct {
	mock.Mock
}

func (m *mockNoncer) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	n, _ := args.Get(0).(uint64)
	return n, args.Error(1)
}

func (m *mockNoncer) CallContext(ctx context.Context, result interface{}, method string, a ...interface{}) error {
	args := m.Called(ctx, result, method, a)
	if args.Get(0) != nil {
		res := result.(*map[string]map[string]map[string]string)
		*res = args.Get(0).(map[string]map[string]map[string]string)
	}

	return args.Error(1)
}

func Test_incrementNonce(t *testing.T) {
	opts := &bind.TransactOpts{From: common.HexToAddress("0x45B9c4798999FFa52e1ff1eFce9d3e45819E4158")}
	gc := &gethClient{
		config: cfg,
	}

	// noncer failed
	mockClient := &MockEthCl{}
	mockClient.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(0), errors.New("error")).Once()
	gc.client = mockClient
	err := gc.setNonce(opts)
	mockClient.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get chain nonce")

	// noncer success
	mockClient.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(1), nil).Once()
	gc.client = mockClient
	err = gc.setNonce(opts)
	mockClient.AssertExpectations(t)
	assert.NoError(t, err)
}
