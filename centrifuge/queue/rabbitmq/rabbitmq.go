package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/ethereum/go-ethereum/log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
)

type Queue struct{
	conn *amqp.Connection
	anchorq amqp.Queue
}

func (q *Queue) Start() {
	// TODO read connection details from configs
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic("Could not connect to rabbitmq")
	}
	q.conn = conn
	ch := openChannel(q)
	// TODO create queues here based on a config if they exist in rabbitmq
	anchorq, err := ch.QueueDeclare(
		"anchoring", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		panic("Could not create anchor queue")
	}
	q.anchorq = anchorq
}

func (q *Queue) Enqueue(queueName string, msg queue.Message) error {
	// TODO get the right queue for the provided name
	ch := openChannel(q)
	defer ch.Close()
	body, err := msg.MarshalBinary()
	if err != nil {
		log.Error("Message could not be marshaled", err)
	}
	err = ch.Publish(
		"",     // exchange
		q.anchorq.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	return err
}

func openChannel(q *Queue) *amqp.Channel {
	ch, err := q.conn.Channel()
	if err != nil {
		panic("Could not create channel to rabbitmq")
	}
	return ch
}

func (q *Queue) Dequeue(queue string) (id string, msg []byte, err error) {
	// TODO get the right queue for the provided name
	ch := openChannel(q)
	defer ch.Close()
	var msgs <-chan amqp.Delivery
	msgs, err = ch.Consume(
		q.anchorq.Name,
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil, // args
	)
	if err != nil {
		return "", nil, err
	}
	amqpMsg := <- msgs

	return amqpMsg.MessageId, amqpMsg.Body, nil
}

func (q *Queue) Delete(queue, id string) error {
	panic("implement me")
}

func (q *Queue) DeleteAll(queue string) error {
	panic("implement me")
}

func (q *Queue) Stop() {
	q.conn.Close()
}

var q queue.Queue

// Get the single available queue connection
func GetQueue() queue.Queue {
	if q == nil {
		q = &Queue{}
		q.Start()
	}
	return q
}

type Worker struct{
	// this worker can only manage a single handler
	messageHandler *queue.MessageHandler
}

func (w *Worker) Start(config *queue.WorkerConfig) {
	// TODO this may need some refactoring to be able for both registry and the code that supplies the handler to be able to start the workers
	// or for those workers that did start already in a pool, this function should not have any effect
	go func() {
		_, msg, err := GetQueue().Dequeue("anchoring") // id is irrelevant for now
		if err != nil {
			log.Error("Error while retrieving messages from queue ", err)
		}
		status, errr := (*w.messageHandler)(msg)
		if errr != nil {
			log.Error("Error while calling message handler ", err, status)
		}
	}()
}

func (w *Worker) AddHandler(handler queue.MessageHandler) {
	w.messageHandler = &handler
}

func (w *Worker) RemoveAllHandlers() {
	w.messageHandler = nil
}

func (w *Worker) Stop() {
	//
}

type WorkerRegistry struct{}

func (workers *WorkerRegistry) Start() {
	// nothing to do here as we initialize workers ondemand
}

func (workers *WorkerRegistry) Get(queueName string) (queue.Worker, error) {
	return &Worker{}, nil
}

func (workers *WorkerRegistry) Stop() {
	panic("implement me")
}

var workerRegistry *WorkerRegistry

func GetWorkerRegistry() queue.WorkerRegistry {
	if (workerRegistry == nil) {
		workerRegistry = &WorkerRegistry{}
	}
	return workerRegistry
}