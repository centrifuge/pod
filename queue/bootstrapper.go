package queue

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
)

// BootstrappedQueuedTasks is a key to tasks that needs to registered with the queue.
const BootstrappedQueuedTasks string = "BootstrappedQueuedTasks"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initiates the queue.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(*config.Configuration)

	// to see how BootstrappedQueuedTasks get populated check usages of InstallQueuedTask
	queuedTasks, ok := context[BootstrappedQueuedTasks]
	if !ok {
		return errors.New("could not find the list of " + BootstrappedQueuedTasks)
	}

	queuedTasksTyped, ok := queuedTasks.([]QueuedTask)
	if !ok {
		return fmt.Errorf("unknown type %T. Required type %T", queuedTasks, []QueuedTask{})
	}

	InitQueue(queuedTasksTyped, cfg.GetNumWorkers(), cfg.GetWorkerWaitTimeMS())
	return nil
}

// InstallQueuedTask adds a queued task to the context so that when the queue initializes it can update it self
// with different tasks types queued in the node
func InstallQueuedTask(context map[string]interface{}, queuedTask QueuedTask) error {
	queuedTasks, ok := context[BootstrappedQueuedTasks]
	if !ok {
		context[BootstrappedQueuedTasks] = []QueuedTask{queuedTask}
		return nil
	}

	queuedTasksTyped, ok := queuedTasks.([]QueuedTask)
	if !ok {
		return errors.New(BootstrappedQueuedTasks + " is of an unexpected type")
	}

	context[BootstrappedQueuedTasks] = append(queuedTasksTyped, queuedTask)
	return nil
}
