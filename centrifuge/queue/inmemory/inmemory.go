package inmemory

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"

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

type InmemoryWorker struct{}

func (iw *InmemoryWorker) Start(config interface{}) {
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

type WorkerConfig struct {
	queueName string
}

func (*WorkerConfig) Start() {
	panic("implement me")
}

func (*WorkerConfig) Get(queueName string) (queue.Worker, error) {
	panic("implement me")
}

func (*WorkerConfig) Stop() {
	panic("implement me")
}

func GetWorkerConfig() WorkerConfig {
	return WorkerConfig{}
}