package anchor

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the AnchorRegistryContract as well as the AnchoringConfirmationTask that depends on it.
// the AnchoringConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	anchorContract, err := getAnchorContract()
	if err != nil {
		return err
	}
	// the following code will add a queued task to the context so that when the queue initializes it can update it self
	// with different tasks types queued in the node
	if queuedTasks, ok := context[queue.BootstrappedQueuedTasks]; ok {
		if queuedTasksTyped, ok := queuedTasks.([]queue.QueuedTask); ok {
			context[queue.BootstrappedQueuedTasks] = append(queuedTasksTyped, createAnchoringConfirmationTask(anchorContract))
			return nil
		} else {
			return errors.New(queue.BootstrappedQueuedTasks + " is of an unexpected type")
		}
	} else {
		context[queue.BootstrappedQueuedTasks] = []queue.QueuedTask{createAnchoringConfirmationTask(anchorContract)}
		return nil
	}
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func createAnchoringConfirmationTask(anchorContract *EthereumAnchorRegistryContract) *AnchoringConfirmationTask {
	return NewAnchoringConfirmationTask(
		&anchorContract.EthereumAnchorRegistryContractFilterer,
		ethereum.DefaultWaitForTransactionMiningContext)
}
