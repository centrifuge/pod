package queue

import (
	"sync"

	"context"
	"errors"
	"time"

	"github.com/centrifuge/gocelery"
	logging "github.com/ipfs/go-log"
)

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

// QueuedTaskType is a task to be queued in the centrifuge node to be completed asynchronously
type QueuedTaskType interface {

	// TaskTypeName of the task
	TaskTypeName() string
}

// QueuedTaskResult represents a result from a queued task execution
type QueuedTaskResult interface {

	// Get the result within a timeout from the queue task execution
	Get(timeout time.Duration) (interface{}, error)
}

// Server represents the queue server currently implemented based on gocelery
type Server struct {
	config    Config
	queue     *gocelery.CeleryClient
	taskTypes []QueuedTaskType
	stop      chan bool
	lock 	  sync.RWMutex
}

// TaskTypeName of the queue server
func (qs *Server) Name() string {
	return "QueueServer"
}

// Start the queue server
func (qs *Server) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
	qs.lock.Lock()
	defer qs.lock.Unlock()
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

	qs.stop = make(chan bool)
	for {
		select {
		case <-qs.stop:
		case <-ctx.Done():
			log.Info("Shutting down Queue server with context done")
			qs.queue.StopWorker()
			log.Info("Queue server stopped")
			return
		}
	}
}

// RegisterTaskType registers a task type on the queue server
func (qs *Server) RegisterTaskType(name string, task interface{}) {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	qs.taskTypes = append(qs.taskTypes, task.(QueuedTaskType))
}

// EnqueueJob enqueues a job on the queue server for the given taskTypeName
func (qs *Server) EnqueueJob(taskTypeName string, params map[string]interface{}) (QueuedTaskResult, error) {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	return qs.queue.DelayKwargs(taskTypeName, params)
}

// Stop force stops the queue server
func (qs *Server) Stop() error {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	if qs.stop != nil {
		qs.stop <- true
	}
	return nil
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
