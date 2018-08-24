package queue

import (
	"sync"

	"github.com/centrifuge/gocelery"
)

var Queue *gocelery.CeleryClient
var queueInit sync.Once

type QueuedTask interface {
	Name() string
	Init() error
}

func InitQueue(tasks []QueuedTask) {
	// TODO do this based on config i.e. numWorkers
	queueInit.Do(func() {
		var err error
		Queue, err = gocelery.NewCeleryClient(gocelery.NewInMemoryBroker(), gocelery.NewInMemoryBackend(), 1)
		for _, task := range tasks {
			task.Init()
		}
		if err != nil {
			panic("Could not initialize the queue")
		}
	})
}
