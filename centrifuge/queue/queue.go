package queue

type EnqueueOptions struct {
	delayMs int
	// to timeout msg receivers or not
	doTimeOut bool
	timeOutMs int
	// max number of retrievals of a message (if the message is a retryable job, this defines the number of retries)
	numRetries int
}

type Header struct {
	CentId   []byte
	TenantId []byte
}

// All queued messages within the Cent node must implement the Message interface
type Message interface {

	Header() *Header

	SerializedMessage() string
}

// Queue interface to be implemented by any queue provider for a Cent node.
// Helps to isolate Cent Node business logic from any specific queue implementation.
type Queue interface {

	// We may need to add a config options object here
	Start()

	// msg can be any deserialized struct, should we change the type to bytes?
	Enqueue(queueName string, msg Message, options *EnqueueOptions) error

	// Dequeue the message but resurface it after the set timeOut
	// (Pull model)
	Dequeue(queue string) (id, msg Message, options *EnqueueOptions, err error)

	// Delete the message with the given id, no resurface afterwards
	Delete(queue, id string) error

	DeleteAll(queue string) error

	Stop()
}

type HandlerStatus int

const (
	SUCCESS HandlerStatus = 0
	ERROR HandlerStatus = 1
	UNKNOWN HandlerStatus = 2
)

// A handler function receives a single message from a queue and handles it after deserializing to proper type.
// Also returns a proper status after the execution.
// Rationale: abstract away the queuing details from business logic. Makes it easier to test the handlers.
type Handler func(msg Message, options *EnqueueOptions) (HandlerStatus, error)

type WorkerConfig struct {
	queueName string
}

// Worker interface is an abstraction over all queue message receivers (go routines).
// It might contain queuing system specific details such as retry logic based on EnqueueOptions.
// Rationale: abstract away the queuing system details from business logic. Makes it easier to test the workers in isolation from business logic.
type Worker interface {

	// We may need to add an config options object here
	Start(config WorkerConfig)

	// Add a handler for the queue that this worker handles
	AddHandler(handler Handler)

	// remove all handlers
	RemoveAllHandlers()

	Stop()
}

// manage the queue workers declared in the system. Holds on to all declared workers in working memory, eg: in a map
type WorkerRegistry interface {

	// Start all the workers.
	Start()

	Get(queueName string) (Worker, error)

	Stop()
}