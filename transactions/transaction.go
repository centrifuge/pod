package transactions

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
)

// Status represents the status of the transaction
type Status uint8

// Status constants
const (
	Success Status = iota
	Failed
	Pending
)

// Log represents a single task in a transaction.
type Log struct {
	Action  string
	Message string
	Time    time.Time
}

// NewLog constructs a new log with action and message
func NewLog(action, message string) Log {
	return Log{
		Action:  action,
		Message: message,
		Time:    time.Now().UTC(),
	}
}

// Transaction contains details of transaction.
type Transaction struct {
	ID          uuid.UUID
	Identity    common.Address
	Description string
	Status      Status
	Logs        []Log
	Metadata    map[string]string
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

// NewTransaction returns a new transaction with a pending state
func NewTransaction(identity common.Address, description string) *Transaction {
	return &Transaction{
		ID:          uuid.Must(uuid.NewV4()),
		Identity:    identity,
		Description: description,
		Status:      Pending,
		Metadata:    make(map[string]string),
	}
}
