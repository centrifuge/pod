// +build unit

package ethereum

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMintingConfirmationTask_ParseKwargs_success(t *testing.T) {
	task := TransactionStatusTask{}
	txHash := "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515"
	jobID := jobs.NewJobID().String()
	did := common.BytesToAddress(utils.RandomSlice(20))

	kwargs := map[string]interface{}{
		jobs.JobIDParam:         jobID,
		TransactionAccountParam: did.String(),
		TransactionTxHashParam:  txHash,
	}

	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Nil(t, err, "parsing should be successful")

	assert.Equal(t, did, task.accountID.ToAddress(), "accountID should be parsed correctly")
	assert.Equal(t, jobID, task.JobID.String(), "jobID should be parsed correctly")
	assert.Equal(t, txHash, task.txHash, "txHash should be parsed correctly")
}

func TestMintingConfirmationTask_ParseKwargsWithEvents_success(t *testing.T) {
	task := TransactionStatusTask{}
	txHash := "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515"
	jobID := jobs.NewJobID().String()
	did := common.BytesToAddress(utils.RandomSlice(20))
	eventName := "IdentityCreated(address)"
	eventIdx := 0

	kwargs := map[string]interface{}{
		jobs.JobIDParam:          jobID,
		TransactionAccountParam:  did.String(),
		TransactionTxHashParam:   txHash,
		TransactionEventName:     eventName,
		TransactionEventValueIdx: eventIdx,
	}

	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Nil(t, err, "parsing should be successful")

	assert.Equal(t, did, task.accountID.ToAddress(), "accountID should be parsed correctly")
	assert.Equal(t, jobID, task.JobID.String(), "jobID should be parsed correctly")
	assert.Equal(t, txHash, task.txHash, "txHash should be parsed correctly")
	assert.Equal(t, eventName, task.eventName, "eventName should be parsed correctly")
	assert.Equal(t, eventIdx, task.eventValueIdx, "eventValueIdx should be parsed correctly")
}

func TestMintingConfirmationTask_ParseKwargs_fail(t *testing.T) {
	task := TransactionStatusTask{}
	eventName := "IdentityCreated(address)"
	tests := []map[string]interface{}{
		{
			jobs.JobIDParam:         jobs.NewJobID().String(),
			TransactionAccountParam: common.BytesToAddress(utils.RandomSlice(20)).String(),
		},
		{
			TransactionAccountParam: common.BytesToAddress(utils.RandomSlice(20)).String(),
			TransactionTxHashParam:  "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
		},
		{
			jobs.JobIDParam:        jobs.NewJobID().String(),
			TransactionTxHashParam: "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
		},
		{
			jobs.JobIDParam:         jobs.NewJobID().String(),
			TransactionAccountParam: common.BytesToAddress(utils.RandomSlice(20)).String(),
			TransactionTxHashParam:  "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
			TransactionEventName:    0,
		},
		{
			jobs.JobIDParam:         jobs.NewJobID().String(),
			TransactionAccountParam: common.BytesToAddress(utils.RandomSlice(20)).String(),
			TransactionTxHashParam:  "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
			TransactionEventName:    eventName,
		},
		{
			jobs.JobIDParam:          jobs.NewJobID().String(),
			TransactionAccountParam:  common.BytesToAddress(utils.RandomSlice(20)).String(),
			TransactionTxHashParam:   "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
			TransactionEventName:     eventName,
			TransactionEventValueIdx: "wrong",
		},
		{
			//empty map

		},
		{
			"dummy": "dummy",
		},
	}

	for i, test := range tests {
		decoded, err := utils.SimulateJSONDecodeForGocelery(test)
		assert.Nil(t, err, "json decode should not thrown an error")
		err = task.ParseKwargs(decoded)
		assert.Error(t, err, "test case %v: parsing should fail", i)
	}
}

func TestGetEventValueFromTransactionReceipt(t *testing.T) {
	eventName := "IdentityCreated(address)"
	eventNameHash := common.BytesToHash(crypto.Keccak256([]byte(eventName)))
	eventValue := []byte{0, 1, 2, 3, 4}
	wrongEvent := "WrongEvent(bytes)"
	eventIdx := 0
	mockClient := &MockEthClient{}

	// Empty event list error
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x1")).Return(&types.Receipt{Status: 1}, nil).Once()
	ethTransTask := NewTransactionStatusTask(200*time.Millisecond, nil, nil, mockClient.TransactionReceipt, nil)
	v, err := ethTransTask.getEventValueFromTransactionReceipt(context.Background(), "0x1", eventName, eventIdx)
	assert.Error(t, err)
	assert.Nil(t, v)

	// Logs missing topics error
	receiptLog := &types.Log{}
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x1")).Return(&types.Receipt{Status: 1, Logs: []*types.Log{receiptLog}}, nil).Once()
	ethTransTask = NewTransactionStatusTask(200*time.Millisecond, nil, nil, mockClient.TransactionReceipt, nil)
	v, err = ethTransTask.getEventValueFromTransactionReceipt(context.Background(), "0x1", eventName, eventIdx)
	assert.Error(t, err)
	assert.Nil(t, v)

	// wrong event filtered
	receiptLog = &types.Log{
		Topics: []common.Hash{
			eventNameHash,
		},
	}
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x1")).Return(&types.Receipt{Status: 1, Logs: []*types.Log{receiptLog}}, nil).Once()
	ethTransTask = NewTransactionStatusTask(200*time.Millisecond, nil, nil, mockClient.TransactionReceipt, nil)
	v, err = ethTransTask.getEventValueFromTransactionReceipt(context.Background(), "0x1", wrongEvent, eventIdx)
	assert.Error(t, err)
	assert.Nil(t, v)

	// wrong event idx filtered
	receiptLog = &types.Log{
		Topics: []common.Hash{
			eventNameHash,
			common.BytesToHash(eventValue),
		},
	}
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x1")).Return(&types.Receipt{Status: 1, Logs: []*types.Log{receiptLog}}, nil).Once()
	ethTransTask = NewTransactionStatusTask(200*time.Millisecond, nil, nil, mockClient.TransactionReceipt, nil)
	v, err = ethTransTask.getEventValueFromTransactionReceipt(context.Background(), "0x1", eventName, 2)
	assert.Error(t, err)
	assert.Nil(t, v)

	// Success
	receiptLog = &types.Log{
		Topics: []common.Hash{
			eventNameHash,
			common.BytesToHash(eventValue),
		},
	}
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x1")).Return(&types.Receipt{Status: 1, Logs: []*types.Log{receiptLog}}, nil).Once()
	ethTransTask = NewTransactionStatusTask(200*time.Millisecond, nil, nil, mockClient.TransactionReceipt, nil)
	v, err = ethTransTask.getEventValueFromTransactionReceipt(context.Background(), "0x1", eventName, eventIdx)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x2, 0x3, 0x4}, v)
}
