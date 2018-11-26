package queue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/centrifuge/gocelery"
	logging "github.com/ipfs/go-log"
)

// Constants are commonly used by all the tasks through kwargs.
const (
	BlockHeightParam string = "BlockHeight"
	TimeoutParam     string = "Timeout"
)

var log = logging.Logger("queue-server")

// Config is an interface for queue specific configurations
type Config interface {

	// GetNumWorkers gets the number of background workers to initiate
	GetNumWorkers() int

	// GetWorkerWaitTime gets the worker wait time for a task to be available while polling
	// increasing this may slow down task execution while reducing it may consume a lot of CPU cycles
	GetWorkerWaitTimeMS() int
}

// TaskType is a task to be queued in the centrifuge node to be completed asynchronously
type TaskType interface {

	// TaskTypeName of the task
	TaskTypeName() string
}

// TaskResult represents a result from a queued task execution
type TaskResult interface {

	// Get the result within a timeout from the queue task execution
	Get(timeout time.Duration) (interface{}, error)
}

// Server represents the queue server currently implemented based on gocelery
type Server struct {
	config    Config
	lock      sync.RWMutex
	queue     *gocelery.CeleryClient
	taskTypes []TaskType
}

// Name of the queue server
func (qs *Server) Name() string {
	return "QueueServer"
}

// Start the queue server
func (qs *Server) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
	qs.lock.Lock()
	var err error
	qs.queue, err = gocelery.NewCeleryClient(
		gocelery.NewInMemoryBroker(),
		gocelery.NewInMemoryBackend(),
		qs.config.GetNumWorkers(),
		qs.config.GetWorkerWaitTimeMS(),
	)
	if err != nil {
		startupErr <- err
	}
	for _, task := range qs.taskTypes {
		qs.queue.Register(task.TaskTypeName(), task)
	}
	// start the workers
	qs.queue.StartWorker()
	qs.lock.Unlock()

	<-ctx.Done()
	log.Info("Shutting down Queue server with context done")
	qs.lock.Lock()
	qs.queue.StopWorker()
	qs.lock.Unlock()
	log.Info("Queue server stopped")
}

// RegisterTaskType registers a task type on the queue server
func (qs *Server) RegisterTaskType(name string, task interface{}) {
	qs.taskTypes = append(qs.taskTypes, task.(TaskType))
}

// EnqueueJob enqueues a job on the queue server for the given taskTypeName
func (qs *Server) EnqueueJob(taskTypeName string, params map[string]interface{}) (TaskResult, error) {
	qs.lock.RLock()
	defer qs.lock.RUnlock()
	if qs.queue == nil {
		return nil, errors.New("queue hasn't been initialised")
	}
	return qs.queue.DelayKwargs(taskTypeName, params)
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
		}
	}
	return 0, errors.New("value can not be parsed")
}
