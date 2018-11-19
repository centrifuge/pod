package queue

import (
	"sync"

	"github.com/centrifuge/gocelery"
	"context"
	"time"
	logging "github.com/ipfs/go-log"
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

// QueueServer represents the queue server currently implemented based on gocelery
type QueueServer struct {
	config    Config
	queue     *gocelery.CeleryClient
	taskTypes []QueuedTaskType
	stop 	  chan bool
}

// TaskTypeName of the queue server
func (qs *QueueServer) Name() string {
	return "QueueServer"
}

// Start the queue server
func (qs *QueueServer) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
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
func (qs *QueueServer) RegisterTaskType(name string, task interface{}) {
	qs.queue.Register(name, task)
}

// EnqueueJob enqueues a job on the queue server for the given taskTypeName
func (qs *QueueServer) EnqueueJob(taskTypeName string, params map[string]interface{}) (QueuedTaskResult, error) {
	return qs.queue.DelayKwargs(taskTypeName, params)
}

// Stop force stops the queue server
func (qs *QueueServer) Stop() error {
	qs.stop <- true
	return nil
}
