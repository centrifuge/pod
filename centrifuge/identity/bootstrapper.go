package identity

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	identityContract, err := getIdentityFactoryContract()
	if err != nil {
		return err
	}
	// the following code will add a queued task to the context so that when the queue initializes it can update it self
	// with different tasks types queued in the node
	if queuedTasks, ok := context[queue.BootstrappedQueuedTasks]; ok {
		if queuedTasksTyped, ok := queuedTasks.([]queue.QueuedTask); ok {
			queuedTasksTyped = append(queuedTasksTyped, createIdRegistrationConfirmationTask(identityContract))
			return nil
		} else {
			return errors.New(queue.BootstrappedQueuedTasks + " is of an unexpected type")
		}
	} else {
		context[queue.BootstrappedQueuedTasks] = []queue.QueuedTask{createIdRegistrationConfirmationTask(identityContract)}
		return nil
	}
}

func createIdRegistrationConfirmationTask(identityContract *EthereumIdentityFactoryContract) *IdRegistrationConfirmationTask {
	return NewIdRegistrationConfirmationTask(
		&identityContract.EthereumIdentityFactoryContractFilterer,
		ethereum.DefaultWaitForTransactionMiningContext)
}
