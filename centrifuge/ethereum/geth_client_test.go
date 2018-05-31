// +build unit

package ethereum_test

import (
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

type MockTransactionInterface interface {
	RegisterTransaction(someVar string, anotherVar string) (tx *types.Transaction, err error)
}

type MockTransactionRequest struct {
	count int
}

func (transactionRequest *MockTransactionRequest) RegisterTransaction(transactionName string, anotherVar string) (tx *types.Transaction, err error) {
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

func TestGetGethTxOpts(t *testing.T) {
	resetMock := testingutils.MockConfigOption("ethereum.maxRetries", 0)
	defer resetMock()

	//invalid input params
	bytes, err := tools.StringToByte32("too short")
	assert.EqualValuesf(t, [32]byte{}, bytes, "Should receive empty byte array if string is not 32 chars")
	assert.Error(t, err, "Should return error on invalid input parameter")

	bytes, err = tools.StringToByte32("")
	assert.EqualValuesf(t, [32]byte{}, bytes, "Should receive empty byte array if string is not 32 chars")
	assert.Error(t, err, "Should return error on invalid input parameter")

	bytes, err = tools.StringToByte32("too long. 12345678901234567890123456789032")
	assert.EqualValuesf(t, [32]byte{}, bytes, "Should receive empty byte array if string is not 32 chars")
	assert.Error(t, err, "Should return error on invalid input parameter")

	//valid input param
	convertThis := "12345678901234567890123456789032"
	bytes, err = tools.StringToByte32(convertThis)
	assert.Nil(t, err, "Should not return error on 32 length string")

	convertedBack, _ := tools.Byte32ToString(bytes)
	assert.EqualValues(t, convertThis, convertedBack, "Converted back value should be the same as original input")
}

func TestInitTransactionWithRetries(t *testing.T) {
	mockRequest := &MockTransactionRequest{}

	// Success at first
	tx, err := ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, "var1", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 1, tx.Nonce(), "Nonce should equal to the one provided")
	assert.EqualValues(t, 1, mockRequest.count, "Transaction Run flag should be true")

	// Failure with non-locking error
	tx, err = ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, "otherError", "var2")
	assert.EqualError(t, err, "Some other error", "Should error out")

	mockRetries := testingutils.MockConfigOption("ethereum.maxRetries", 10)
	defer mockRetries()

	mockRequest.count = 0
	// Failure and timeout with locking error
	tx, err = ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, "optimisticLockingTimeout", "var2")
	assert.EqualError(t, err, ethereum.TransactionUnderpriced, "Should error out")
	assert.EqualValues(t, 10, mockRequest.count, "Retries should be equal")

	mockRequest.count = 0
	// Success after locking race condition overcome
	tx, err = ethereum.SubmitTransactionWithRetries(mockRequest.RegisterTransaction, "optimisticLockingEventualSuccess", "var2")
	assert.Nil(t, err, "Should not error out")
	assert.EqualValues(t, 3, mockRequest.count, "Retries should be equal")
}
