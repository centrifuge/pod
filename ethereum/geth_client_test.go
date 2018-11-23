// +build unit

package ethereum

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	cfg.Set("ethereum.txPoolAccessEnabled", false)
	cfg.Set("ethereum.intervalRetry", time.Millisecond*100)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

type MockTransactionRequest struct {
	count int
}

func (transactionRequest *MockTransactionRequest) RegisterTransaction(opts *bind.TransactOpts, transactionName string, anotherVar string) (tx *types.Transaction, err error) {
	transactionRequest.count++
	if transactionName == "otherError" {
		err = errors.Wrap("Some other error", 1)
	} else if transactionName == "optimisticLockingTimeout" {
		err = errors.Wrap(transactionUnderpriced, 1)
	} else if transactionName == "optimisticLockingEventualSuccess" {
		if transactionRequest.count < 3 {
			err = errors.Wrap(transactionUnderpriced, 1)
		}
	}

	return types.NewTransaction(1, common.Address{}, nil, 0, nil, nil), err
}

func TestInitTransactionWithRetries(t *testing.T) {
	mockRequest := &MockTransactionRequest{}

	gc := &gethClient{
		accounts: make(map[string]*bind.TransactOpts),
		accMu:    sync.Mutex{},
		txMu:     sync.Mutex{},
		config:   cfg,
	}

	SetClient(gc)

	// Success at first
	tx, err := gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "var1", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 1, tx.Nonce(), "Nonce should equal to the one provided")
	assert.EqualValues(t, 1, mockRequest.count, "Transaction Run flag should be true")

	// Failure with non-locking error
	tx, err = gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "otherError", "var2")
	assert.EqualError(t, err, "Some other error", "Should error out")

	mockRetries := testingutils.MockConfigOption(cfg, "ethereum.maxRetries", 10)
	defer mockRetries()

	mockRequest.count = 0
	// Failure and timeout with locking error
	tx, err = gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "optimisticLockingTimeout", "var2")
	assert.Contains(t, err.Error(), transactionUnderpriced, "Should error out")
	assert.EqualValues(t, 10, mockRequest.count, "Retries should be equal")

	mockRequest.count = 0
	// Success after locking race condition overcome
	tx, err = gc.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "optimisticLockingEventualSuccess", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 3, mockRequest.count, "Retries should be equal")
}

func TestGetGethCallOpts(t *testing.T) {
	opts, cancel := GetClient().GetGethCallOpts()
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

	// txpool access disabled
	err := gc.incrementNonce(opts, false, nil, nil)
	assert.Nil(t, err)
	assert.Nil(t, opts.Nonce)

	// noncer failed
	n := new(mockNoncer)
	n.On("PendingNonceAt", mock.Anything, opts.From).Return(0, fmt.Errorf("error")).Once()
	err = gc.incrementNonce(opts, true, n, nil)
	n.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get chain nonce")

	// rpc call failed
	n = new(mockNoncer)
	n.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(100), nil).Once()
	n.On("CallContext", mock.Anything, mock.Anything, "txpool_inspect", mock.Anything).Return(nil, fmt.Errorf("error")).Once()
	err = gc.incrementNonce(opts, true, n, n)
	n.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get txpool data")
	assert.Equal(t, "100", opts.Nonce.String())

	// no pending txns in the tx pool
	n = new(mockNoncer)
	n.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(1000), nil).Once()
	n.On("CallContext", mock.Anything, mock.Anything, "txpool_inspect", mock.Anything).Return(nil, nil).Once()
	err = gc.incrementNonce(opts, true, n, n)
	n.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, "1000", opts.Nonce.String())

	// bad result
	var res = map[string]map[string]map[string]string{
		"pending": {
			opts.From.String(): {
				"abc": "0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas",
			},
		},
	}
	n = new(mockNoncer)
	n.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(1000), nil).Once()
	n.On("CallContext", mock.Anything, mock.Anything, "txpool_inspect", mock.Anything).Return(res, nil).Once()
	err = gc.incrementNonce(opts, true, n, n)
	n.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to convert nonce")

	// higher nonce in txpool
	res = map[string]map[string]map[string]string{
		"pending": {
			opts.From.String(): {
				"1000": "0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas",
				"1001": "0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas",
				"1002": "0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas",
				"1003": "0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas",
			},
		},
	}
	n = new(mockNoncer)
	n.On("PendingNonceAt", mock.Anything, opts.From).Return(uint64(1000), nil).Once()
	n.On("CallContext", mock.Anything, mock.Anything, "txpool_inspect", mock.Anything).Return(res, nil).Once()
	err = gc.incrementNonce(opts, true, n, n)
	n.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, "1004", opts.Nonce.String())
}
