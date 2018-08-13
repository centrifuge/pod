package inmemory

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"

///////////////////////////////////////////////////////////////////
// An in memory implementation of the queue.Queue interface
///////////////////////////////////////////////////////////////////

type Queue struct{}

func (Queue) Start() {
	panic("implement me")
}

func (iq *Queue) Enqueue(queueName string, msg string, options *queue.EnqueueOptions) error {
	panic("implement me")
}

func (iq *Queue) Dequeue(queue string) (id, msg string, options *queue.EnqueueOptions, err error) {
	panic("implement me")
}

func (iq *Queue) Delete(queue, id string) error {
	panic("implement me")
}

func (iq *Queue) DeleteAll(queue string) error {
	panic("implement me")
}

func (iq *Queue) Stop() {
	panic("implement me")
}

///////////////////////////////////////////////////////////////////
// An inmemory queue based implementation of queue.Worker interface
///////////////////////////////////////////////////////////////////

type Worker struct{}

func (iw *Worker) Start(config queue.WorkerConfig) {
	panic("implement me")
}

func (iw *Worker) AddHandler(handler queue.MessageHandler) {
	panic("implement me")
}

func (iw *Worker) RemoveAllHandlers() {
	panic("implement me")
}

func (iw *Worker) Stop() {
	panic("implement me")
}

///////////////////////////////////////////////////////////////////
// An implementation of queue.WorkerRegistry for InmemoryWorkers
//////////////////////////////////////////////////////////////////

type WorkerRegistry struct{}

func (WorkerRegistry) Start() {
	panic("implement me")
}

func (WorkerRegistry) Get(queueName string) (queue.Worker, error) {
	panic("implement me")
}

func (WorkerRegistry) Stop() {
	panic("implement me")
}

func GetWorkerRegistry() queue.WorkerRegistry {
	// TODO modify this to store the registry inmemory and return a ref to that
	return WorkerRegistry{}
}