package queue

import (
	"sync"

	"errors"
	"time"

	"github.com/centrifuge/gocelery"
)

// Constant key params which are passed as kwargs to queue.
const (
	BlockHeightParam string = "BlockHeight"
	TimeoutParam     string = "Timeout"
)

// Queue is the initialise celeryClient
var Queue *gocelery.CeleryClient
var queueInit sync.Once

// QueuedTask is a task to be queued in the centrifuge node to be completed asynchronously
type QueuedTask interface {
	Name() string
	Init() error
}

// InitQueue initialises the queue.
func InitQueue(tasks []QueuedTask, numWorkers, workerWaitTime int) {
	queueInit.Do(func() {
		var err error
		Queue, err = gocelery.NewCeleryClient(
			gocelery.NewInMemoryBroker(),
			gocelery.NewInMemoryBackend(),
			numWorkers,
			workerWaitTime,
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

// StopQueue stops the current queue client.
func StopQueue() {
	Queue.StopWorker()
}

// GetDuration parses key parameter to time.Duration type
func GetDuration(key interface{}) (time.Duration, error) {
	f64, ok := key.(float64)
	if !ok {
		return time.Duration(0), errors.New("could not parse interface to float64")
	}
	return time.Duration(f64), nil
}

// ParseBlockHeight parses blockHeight interface param to uint64
func ParseBlockHeight(valMap map[string]interface{}) (uint64, error) {
	bhi, ok := valMap[BlockHeightParam]
	if !ok {
		return 0, errors.New("value can not be parsed")
	}

	bhf, ok := bhi.(float64)
	if !ok {
		return 0, errors.New("value can not be parsed")
	}

	return uint64(bhf), nil
}
