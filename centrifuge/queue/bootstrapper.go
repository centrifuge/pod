package queue

import "errors"

const BootstrappedQueuedTasks string = "BootstrappedQueuedTasks"

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if queuedTasks, ok := context[BootstrappedQueuedTasks]; ok {
		if queuedTasksTyped, ok := queuedTasks.([]QueuedTask); ok {
			InitQueue(queuedTasksTyped)
			return nil
		}
	}
	return errors.New("could not find the list of " + BootstrappedQueuedTasks)
}
