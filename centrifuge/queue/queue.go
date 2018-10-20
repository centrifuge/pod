package queue

import (
	"sync"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/gocelery"
)

var Queue *gocelery.CeleryClient
var queueInit sync.Once

// QueuedTask is a task to be queued in the centrifuge node to be completed asynchronously
type QueuedTask interface {
	Name() string
	Init() error
}

func InitQueue(tasks []QueuedTask) {
	queueInit.Do(func() {
		var err error
		Queue, err = gocelery.NewCeleryClient(
			gocelery.NewInMemoryBroker(),
			gocelery.NewInMemoryBackend(),
			config.Config.GetNumWorkers(),
			config.Config.GetWorkerWaitTimeMS(),
		)
		if err != nil {
			panic("Could not initialize the queue")
		}
		for _, task := range tasks {
			task.Init()
		}
		Queue.StartWorker()
	})
}

func StopQueue() {
	Queue.StopWorker()
}
