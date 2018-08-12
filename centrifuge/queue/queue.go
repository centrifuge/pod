package queue

type EnqueueOptions struct {
	delayMs int
	// to timeout msg receivers or not
	doTimeOut bool
	timeOutMs int
	// max number of retrievals of a message (if the message is a retryable job, this defines the number of retries)
	numRetries int
}

// Queue interface to be implemented by any queue provider for a Cent node
type Queue interface {

	// We may need to add an config options object here
	Start()

	// msg can be any deserialized struct, should we change the type to bytes?
	Enqueue(queueName string, msg string, options *EnqueueOptions)

	// Dequeue the message but resurface it after the set timeOut
	// (Pull model)
	Dequeue(queue string) (id, msg string, options *EnqueueOptions)

	// Delete the message with the given id, no resurface afterwards
	Delete(queue, id string)

	DeleteAll(queue string)

	Stop()
}

type HandlerStatus int

const (
	SUCCESS HandlerStatus = 0
	ERROR HandlerStatus = 1
	UNKNOWN HandlerStatus = 2
)

// A handler function receives a single message from a queue and handles it(after deserializing to proper type), returning a proper status after the execution
// Rationale: abstract away the queuing details from business logic
type Handler func(msg string, options *EnqueueOptions) HandlerStatus

// Worker interface is an abstraction over all queue message receivers (go routines).
// It might contain queuing system specific details such as retry logic based on EnqueueOptions.
// Rationale: abstract away the queuing system details from business logic
type Worker interface {

	// We may need to add an config options object here
	Start()

	AddHandler(queueName string, handler Handler)

	RemoveHandlers(queueName string)

	Stop()
}