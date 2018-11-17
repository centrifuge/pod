package queue

import (
	"sync"

	"errors"
	"time"

	"github.com/centrifuge/gocelery"
)

const (
	BlockHeightParam string = "BlockHeight"
	TimeoutParam     string = "Timeout"
)

var Queue *gocelery.CeleryClient
var queueInit sync.Once

// QueuedTask is a task to be queued in the centrifuge node to be completed asynchronously
type QueuedTask interface {
	Name() string
	Init() error
}

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

func StopQueue() {
	Queue.StopWorker()
}

// GetDuration parses key parameter to time.Duration type
func GetDuration(key interface{}) (time.Duration, error) {
	f64, ok := key.(float64)
	if !ok {
		return time.Duration(0), errors.New("Could not parse interface to float64")
	}
	return time.Duration(f64), nil
}

// ParseBlockHeight parses blockHeight interface param to uint64
func ParseBlockHeight(valMap map[string]interface{}) (uint64, error) {
	if bhi, ok := valMap[BlockHeightParam]; ok {
		bhf, ok := bhi.(float64)
		if ok {
			return uint64(bhf), nil
		} else {
			return 0, errors.New("value can not be parsed")
		}
	}
	return 0, errors.New("value can not be parsed")
}
