// +build unit

package ethereum

import (
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestWaitForEventTask_ParseKwargs(t *testing.T) {
	task := new(WaitForEventTask)
	jobsID := jobs.NewJobID().String()
	accountID := common.BytesToAddress(utils.RandomSlice(20))
	eventSign := "AssetStored(bytes32)"
	// missing from block
	kwargs := map[string]interface{}{
		jobs.JobIDParam:           jobsID,
		TransactionAccountParam:   accountID.String(),
		WaitForEventNameSignature: eventSign,
	}
	checkKwargs(t, task, kwargs, false)

	// missing address
	from := big.NewInt(1000)
	kwargs = map[string]interface{}{
		jobs.JobIDParam:           jobsID,
		TransactionAccountParam:   accountID.String(),
		WaitForEventFromBlock:     hexutil.EncodeBig(from),
		WaitForEventNameSignature: eventSign,
	}
	checkKwargs(t, task, kwargs, false)

	// missing topic
	address := common.BytesToAddress(utils.RandomSlice(20))
	kwargs = map[string]interface{}{
		jobs.JobIDParam:           jobsID,
		TransactionAccountParam:   accountID.String(),
		WaitForEventFromBlock:     hexutil.EncodeBig(from),
		WaitForEventAddress:       address.Hex(),
		WaitForEventNameSignature: eventSign,
	}
	checkKwargs(t, task, kwargs, false)

	// actual query
	topic := common.BytesToHash(utils.RandomSlice(32))
	kwargs = map[string]interface{}{
		jobs.JobIDParam:           jobsID,
		TransactionAccountParam:   accountID.String(),
		WaitForEventFromBlock:     hexutil.EncodeBig(from),
		WaitForEventAddress:       address.Hex(),
		WaitForEventTopic:         topic.Hex(),
		WaitForEventNameSignature: eventSign,
	}
	checkKwargs(t, task, kwargs, true)
	eq := ethereum.FilterQuery{
		FromBlock: from,
		Addresses: []common.Address{address},
		Topics: [][]common.Hash{
			{common.BytesToHash(crypto.Keccak256([]byte(eventSign)))},
			{topic},
		},
	}
	assert.Equal(t, eq, task.query)
}

func checkKwargs(t *testing.T, task *WaitForEventTask, kwargs map[string]interface{}, success bool) {
	skwargs, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.NoError(t, err)
	err = task.ParseKwargs(skwargs)
	if success {
		assert.NoError(t, err)
		return
	}
	assert.Error(t, err)
}

// func TestWaitForEventTask_RunTask(t *testing.T) {
// 	task := new(WaitForEventTask)
// 	jm := testingjobs.MockJobManager{}
// 	jm.On("UpdateTaskStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
// 	task.BaseTask = jobsv1.BaseTask{
// 		JobManager: jm,
// 	}
// 	task.ethContextInitializer = func() (ctx context.Context, cancelFunc context.CancelFunc) {
// 		t := time.Now().Add(5 * time.Second)
// 		return context.WithDeadline(context.Background(), t)
// 	}
//
// 	// deadline exceeded
// 	task.filterLogsFunc = func(ctx context.Context, query ethereum.FilterQuery) (logs []types.Log, err error) {
// 		return nil, context.DeadlineExceeded
// 	}
// 	_, err := task.RunTask()
// 	assert.Equal(t, err, gocelery.ErrTaskRetryable)
//
// 	// non retryable
// 	task.filterLogsFunc = func(ctx context.Context, query ethereum.FilterQuery) (logs []types.Log, err error) {
// 		return nil, errors.New("random error")
// 	}
// 	_, err = task.RunTask()
// 	assert.Error(t, err)
// 	assert.Equal(t, err.Error(), "[random error]")
//
// 	// no logs
// 	task.filterLogsFunc = func(ctx context.Context, query ethereum.FilterQuery) (logs []types.Log, err error) {
// 		return nil, nil
// 	}
// 	_, err = task.RunTask()
// 	assert.Error(t, err)
// 	assert.Equal(t, err, gocelery.ErrTaskRetryable)
//
// 	// success
// 	task.filterLogsFunc = func(ctx context.Context, query ethereum.FilterQuery) (logs []types.Log, err error) {
// 		return []types.Log{
// 			{},
// 		}, nil
// 	}
// 	_, err = task.RunTask()
// 	assert.NoError(t, err)
// 	jm.AssertExpectations(t)
// }
