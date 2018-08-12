package inmemory

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"

///////////////////////////////////////////////////////////////////
// An inmemory implementation of the queue.Queue interface
///////////////////////////////////////////////////////////////////

type InmemoryQueue struct{}

func (InmemoryQueue) Start() {
	panic("implement me")
}

func (iq *InmemoryQueue) Enqueue(queueName string, msg string, options *queue.EnqueueOptions) error {
	panic("implement me")
}

func (iq *InmemoryQueue) Dequeue(queue string) (id, msg string, options *queue.EnqueueOptions, err error) {
	panic("implement me")
}

func (iq *InmemoryQueue) Delete(queue, id string) error {
	panic("implement me")
}

func (iq *InmemoryQueue) DeleteAll(queue string) error {
	panic("implement me")
}

func (iq *InmemoryQueue) Stop() {
	panic("implement me")
}

///////////////////////////////////////////////////////////////////
// An inmemory queue based implementation of queue.Worker interface
///////////////////////////////////////////////////////////////////

type InmemoryWorker struct{}

func (iw *InmemoryWorker) Start(config queue.WorkerConfig) {
	panic("implement me")
}

func (iw *InmemoryWorker) AddHandler(handler queue.Handler) {
	panic("implement me")
}

func (iw *InmemoryWorker) RemoveAllHandlers() {
	panic("implement me")
}

func (iw *InmemoryWorker) Stop() {
	panic("implement me")
}
///////////////////////////////////////////////////////////////////
// An implementation of queue.WorkerRegistry for InmemoryWorkers
//////////////////////////////////////////////////////////////////


type InmemoryWorkerRegistry struct{}

func (InmemoryWorkerRegistry) Start() {
	panic("implement me")
}

func (InmemoryWorkerRegistry) Get(queueName string) (queue.Worker, error) {
	panic("implement me")
}

func (InmemoryWorkerRegistry) Stop() {
	panic("implement me")
}

func GetWorkerRegistry() queue.WorkerRegistry {
	// TODO modify this to store the registry inmemory and return a ref to that
	return InmemoryWorkerRegistry{}
}