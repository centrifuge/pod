package ethereum

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ETHWaitForEvent is the name of the task
	ETHWaitForEvent = "EthWaitForEventTask"

	// WaitForEventFromBlock is the key for the from block
	WaitForEventFromBlock = "WaitForEventFromBlock"

	// WaitForEventAddress is the key for the address
	WaitForEventAddress = "WaitForEventAddress"

	// WaitForEventNameSignature is the key for the name and signature of the event
	WaitForEventNameSignature = "WaitForEventNameSignature"

	// WaitForEventTopic is the key for the topic
	WaitForEventTopic = "WaitForEventTopic"
)

// WaitForEventTask holds from block, addresses of the contract to listen events from, topics of the event
// the task assumes to block to be the latest
type WaitForEventTask struct {
	jobsv1.BaseTask
	accountID             identity.DID
	query                 ethereum.FilterQuery
	filterLogsFunc        func(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc)
}

// Copy copies the state of the task into a new task
func (t *WaitForEventTask) Copy() (gocelery.CeleryTask, error) {
	return &WaitForEventTask{
		BaseTask:              jobsv1.BaseTask{JobManager: t.JobManager},
		filterLogsFunc:        t.filterLogsFunc,
		ethContextInitializer: t.ethContextInitializer,
	}, nil
}

// ParseKwargs parses the kwargs into a query.
func (t *WaitForEventTask) ParseKwargs(kwargs map[string]interface{}) error {
	err := t.ParseJobID(t.TaskTypeName(), kwargs)
	if err != nil {
		return err
	}

	accountID, ok := kwargs[TransactionAccountParam].(string)
	if !ok {
		return errors.NewTypedError(ErrEthTransaction, errors.New("missing account ID"))
	}

	fb, ok := kwargs[WaitForEventFromBlock].(string)
	if !ok {
		return errors.New("failed to decode from block: %v", kwargs[WaitForEventFromBlock])
	}

	address, ok := kwargs[WaitForEventAddress].(string)
	if !ok {
		return errors.New("failed to decode address: %v", kwargs[WaitForEventAddress])
	}

	eventSignature, ok := kwargs[WaitForEventNameSignature].(string)
	if !ok {
		return errors.New("failed to decode event name signature: %v", kwargs[WaitForEventNameSignature])
	}

	topic, ok := kwargs[WaitForEventTopic].(string)
	if !ok {
		return errors.New("failed to decode topic: %v", kwargs[WaitForEventTopic])
	}

	ehash := common.BytesToHash(crypto.Keccak256([]byte(eventSignature)))
	t.accountID = identity.DID(common.HexToAddress(accountID))
	t.query = ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: hexutil.MustDecodeBig(fb),
		ToBlock:   nil,
		Addresses: []common.Address{common.HexToAddress(address)},
		Topics: [][]common.Hash{
			{ehash},
			{common.HexToHash(topic)},
		},
	}

	return nil
}

// RunTask runs the task of fetching the logs.
func (t *WaitForEventTask) RunTask() (res interface{}, err error) {
	var jobValue *jobs.JobValue
	ctx, cancelFunc := t.ethContextInitializer()
	defer func() {
		err = t.UpdateJobWithValue(t.accountID, t.TaskTypeName(), err, jobValue)
	}()
	defer cancelFunc()

	logs, err := t.filterLogsFunc(ctx, t.query)
	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, gocelery.ErrTaskRetryable
		}

		return nil, err
	}

	if len(logs) == 0 {
		// no logs found, wait for another round
		return nil, gocelery.ErrTaskRetryable
	}

	return nil, nil
}

// TaskTypeName returns EthWaitForEvent
func (t *WaitForEventTask) TaskTypeName() string {
	return ETHWaitForEvent
}

// NewWaitEventTask returns a new wait event task for registration
func NewWaitEventTask(
	jobsManager jobs.Manager,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
	filterLogsFunc func(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error),
) *WaitForEventTask {
	return &WaitForEventTask{
		BaseTask:              jobsv1.BaseTask{JobManager: jobsManager},
		filterLogsFunc:        filterLogsFunc,
		ethContextInitializer: ethContextInitializer,
	}
}

func initWaitForEventTask(
	tq queue.TaskQueuer,
	accountID identity.DID,
	jobID jobs.JobID,
	eventSignature string,
	fromBlock *big.Int, address common.Address, topic common.Hash,
) (queue.TaskResult, error) {
	params := map[string]interface{}{
		jobs.JobIDParam:           jobID.String(),
		TransactionAccountParam:   accountID.String(),
		WaitForEventFromBlock:     hexutil.EncodeBig(fromBlock),
		WaitForEventAddress:       address.Hex(),
		WaitForEventTopic:         topic.Hex(),
		WaitForEventNameSignature: eventSignature,
	}

	return tq.EnqueueJob(ETHWaitForEvent, params)
}

// CreateWaitForEventJob creates a job for waiting for event from ethereum
func CreateWaitForEventJob(
	parentCtx context.Context,
	jobsMan jobs.Manager,
	tq queue.TaskQueuer,
	self identity.DID,
	jobID jobs.JobID,
	eventSignature string,
	fromBlock *big.Int, address common.Address, topic common.Hash) (jobs.JobID, chan error, error) {
	jobID, done, err := jobsMan.ExecuteWithinJob(contextutil.Copy(parentCtx), self, jobID, "Waiting for Event from Ethereum", func(accountID identity.DID, jobID jobs.JobID, jobsMan jobs.Manager, errChan chan<- error) {
		tr, err := initWaitForEventTask(tq, accountID, jobID, eventSignature, fromBlock, address, topic)
		if err != nil {
			errChan <- err
			return
		}
		_, err = tr.Get(jobsMan.GetDefaultTaskTimeout())
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	})
	return jobID, done, err
}
