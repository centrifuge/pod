package transactions

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
)

// Status represents the status of the transaction
type Status string

const (
	// Status constants

	// Success is the success status for a transaction or a task
	Success Status = "success"
	// Failed is the failed status for a transaction or a task
	Failed Status = "failed"
	// Pending is the pending status for a transaction or a task
	Pending Status = "pending"

	// TxIDParam maps transaction ID in the kwargs.
	TxIDParam = "transactionID"

	// BootstrappedRepo is the key mapped to transactions.Repository.
	BootstrappedRepo = "BootstrappedRepo"

	// BootstrappedService is the key to mapped transactions.Manager
	BootstrappedService = "BootstrappedService"
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

// TxID is a centrifuge transaction ID. Internally represented by a UUID. Externally visible as a byte slice or a hex encoded string.
type TxID uuid.UUID

// NewTxID creates a new TxID
func NewTxID() TxID {
	u := uuid.Must(uuid.NewV4())
	return TxID(u)
}

// FromString tries to convert the given hex string txID into a type TxID
func FromString(tIDHex string) (TxID, error) {
	tidBytes, err := hexutil.Decode(tIDHex)
	if err != nil {
		return NilTxID(), err
	}
	u, err := uuid.FromBytes(tidBytes)
	if err != nil {
		return NilTxID(), err
	}
	return TxID(u), nil
}

// NilTxID returns a nil TxID
func NilTxID() TxID {
	return TxID(uuid.Nil)
}

// String marshals a TxID to its hex string form
func (t TxID) String() string {
	return hexutil.Encode(t[:])
}

// Bytes returns the byte slice representation of the TxID
func (t TxID) Bytes() []byte {
	return uuid.UUID(t).Bytes()
}

// TxIDEqual checks if given two TxIDs are equal
func TxIDEqual(t1 TxID, t2 TxID) bool {
	u1 := uuid.UUID(t1)
	u2 := uuid.UUID(t2)
	return uuid.Equal(u1, u2)
}

// Transaction contains details of transaction.
type Transaction struct {
	ID          TxID
	DID         identity.DID
	Description string

	// Status is the overall status of the transaction
	Status Status

	// TaskStatus tracks the status of individual tasks running in the system for this transaction
	TaskStatus map[string]Status

	// Logs are transaction log messages
	Logs      []Log
	CreatedAt time.Time

	// Values retrieved from events
	Values map[string]TXValue
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
func NewTransaction(identity identity.DID, description string) *Transaction {
	return &Transaction{
		ID:          NewTxID(),
		DID:         identity,
		Description: description,
		Status:      Pending,
		TaskStatus:  make(map[string]Status),
		CreatedAt:   time.Now().UTC(),
		Values:      make(map[string]TXValue),
	}
}

// TXValue holds the key and value filtered by the transaction
type TXValue struct {
	Key    string
	KeyIdx int
	Value  []byte
}

// Config is the config interface for transactions package
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
}

// Manager is a manager for centrifuge transactions.
type Manager interface {
	// ExecuteWithinTX executes the given unit of work within a transaction
	ExecuteWithinTX(ctx context.Context, accountID identity.DID, existingTxID TxID, desc string, work func(accountID identity.DID, txID TxID, txMan Manager, err chan<- error)) (txID TxID, done chan bool, err error)
	GetTransaction(accountID identity.DID, id TxID) (*Transaction, error)
	UpdateTransactionWithValue(accountID identity.DID, id TxID, key string, value []byte) error
	UpdateTaskStatus(accountID identity.DID, id TxID, status Status, taskName, message string) error
	GetTransactionStatus(accountID identity.DID, id TxID) (*transactionspb.TransactionStatusResponse, error)
	WaitForTransaction(accountID identity.DID, txID TxID) error
	GetDefaultTaskTimeout() time.Duration
}

// Repository can be implemented by a type that handles storage for transactions.
type Repository interface {
	Get(cid identity.DID, id TxID) (*Transaction, error)
	Save(transaction *Transaction) error
}
