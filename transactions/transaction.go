package transactions

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/satori/go.uuid"
)

// Status represents the status of the transaction
type Status string

// Status constants
const (
	Success Status = "success"
	Failed  Status = "failed"
	Pending Status = "pending"
)

// Log represents a single task in a transaction.
type Log struct {
	Action    string
	Message   string
	CreatedAt time.Time
}

// NewLog constructs a new log with action and message
func NewLog(action, message string) Log {
	return Log{
		Action:    action,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
}

// Transaction contains details of transaction.
type Transaction struct {
	ID          uuid.UUID
	DID         identity.DID
	Description string

	// Status is the overall status of the transaction
	Status Status

	// TaskStatus tracks the status of individual tasks running in the system for this transaction
	TaskStatus map[string]Status

	// Logs are transaction log messages
	Logs      []Log
	CreatedAt time.Time
}

// JSON returns json marshaled transaction.
func (t *Transaction) JSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON loads the data into transaction.
func (t *Transaction) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// Type returns the reflect.Type of the transaction.
func (t *Transaction) Type() reflect.Type {
	return reflect.TypeOf(t)
}

// newTransaction returns a new transaction with a pending state
func newTransaction(identity identity.DID, description string) *Transaction {
	return &Transaction{
		ID:          uuid.Must(uuid.NewV4()),
		DID:         identity,
		Description: description,
		Status:      Pending,
		TaskStatus:  make(map[string]Status),
		CreatedAt:   time.Now().UTC(),
	}
}
