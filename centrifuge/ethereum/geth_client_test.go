// +build unit

package ethereum_test

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"sync"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	config.Config.V.Set("ethereum.txPoolAccessEnabled", false)
	config.Config.V.Set("ethereum.intervalRetry", time.Millisecond*100)

	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

type MockTransactionInterface interface {
	RegisterTransaction(opts *bind.TransactOpts, someVar string, anotherVar string) (tx *types.Transaction, err error)
}

type MockTransactionRequest struct {
	count int
}

func (transactionRequest *MockTransactionRequest) RegisterTransaction(opts *bind.TransactOpts, transactionName string, anotherVar string) (tx *types.Transaction, err error) {
	transactionRequest.count++
	if transactionName == "otherError" {
		err = errors.Wrap("Some other error", 1)
	} else if transactionName == "optimisticLockingTimeout" {
		err = errors.Wrap(ethereum.TransactionUnderpriced, 1)
	} else if transactionName == "optimisticLockingEventualSuccess" {
		if transactionRequest.count < 3 {
			err = errors.Wrap(ethereum.TransactionUnderpriced, 1)
		}
	}
	tx = types.NewTransaction(1, common.Address{}, nil, 0, nil, nil)
	return
}

func TestInitTransactionWithRetries(t *testing.T) {
	mockRequest := &MockTransactionRequest{}

	gc := ethereum.GethClient{nil, nil, nil, &sync.Mutex{}, make(map[string]*bind.TransactOpts)}
	ethereum.SetConnection(gc)

	// Success at first
	tx, err := ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "var1", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 1, tx.Nonce(), "Nonce should equal to the one provided")
	assert.EqualValues(t, 1, mockRequest.count, "Transaction Run flag should be true")

	// Failure with non-locking error
	tx, err = ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "otherError", "var2")
	assert.EqualError(t, err, "Some other error", "Should error out")

	mockRetries := testingutils.MockConfigOption("ethereum.maxRetries", 10)
	defer mockRetries()

	mockRequest.count = 0
	// Failure and timeout with locking error
	tx, err = ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "optimisticLockingTimeout", "var2")
	assert.EqualError(t, err, ethereum.TransactionUnderpriced, "Should error out")
	assert.EqualValues(t, 10, mockRequest.count, "Retries should be equal")

	mockRequest.count = 0
	// Success after locking race condition overcome
	tx, err = ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, &bind.TransactOpts{}, "optimisticLockingEventualSuccess", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 3, mockRequest.count, "Retries should be equal")
}

func TestCalculateIncrement(t *testing.T) {
	strAddress := "0x45B9c4798999FFa52e1ff1eFce9d3e45819E4158"
	var input = map[string]map[string]map[string][]string{
		"pending": {
			strAddress: {
				"0": {"0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas"},
				"1": {"0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas"},
				"2": {"0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas"},
				"3": {"0x958c1fa64b34db746925c6f8a3dd81128e40355e: 1051546810000000000 wei + 90000 × 20000000000 gas"},
			},
		},
	}

	opts := &bind.TransactOpts{From: common.HexToAddress(strAddress)}

	// OnChain Transaction Count is behind local tx pool
	chainNonce := 3
	expectedNonce := 4
	ethereum.CalculateIncrement(uint64(chainNonce), input, opts)
	assert.Equal(t, big.NewInt(int64(expectedNonce)), opts.Nonce, "Nonce should match expected value")

	// OnChain Transaction Count is ahead of local tx pool, means that txs will get stuck forever (Rare case)
	chainNonce = 4
	expectedNonce = 4
	ethereum.CalculateIncrement(uint64(chainNonce), input, opts)
	assert.Equal(t, big.NewInt(int64(expectedNonce)), opts.Nonce, "Nonce should match expected value")
}
