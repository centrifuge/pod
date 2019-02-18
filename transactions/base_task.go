package transactions

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/gocelery"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

var log = logging.Logger("transaction")

const (
	// TxIDParam maps transaction ID in the kwargs.
	TxIDParam = "transactionID"
)

// BaseTask holds the required details and helper functions for tasks to update transactions.
// should be embedded into the task
type BaseTask struct {
	TxID uuid.UUID

	// state
	TxManager Manager
}

// ParseTransactionID parses txID.
func (b *BaseTask) ParseTransactionID(taskTypeName string, kwargs map[string]interface{}) error {
	txID, ok := kwargs[TxIDParam].(string)
	if !ok {
		return errors.New("missing transaction ID")
	}

	var err error
	b.TxID, err = uuid.FromString(txID)
	if err != nil {
		return errors.New("invalid transaction ID")
	}

	log.Infof("Task %s parsed for tx: %s\n", taskTypeName, b.TxID)
	return nil
}

// UpdateTransaction add a new log and updates the status of the transaction based on the error.
func (b *BaseTask) UpdateTransaction(accountID identity.CentID, taskTypeName string, err error) error {
	if err == gocelery.ErrTaskRetryable {
		return err
	}

	// TODO this TaskStatus map update assumes that a single transaction has only one execution of a certain task type, which can be wrong, use the taskID or another unique identifier instead.
	if err != nil {
		log.Errorf("Task %s failed for transaction: %v\n", taskTypeName, b.TxID.String())
		return errors.AppendError(err, b.TxManager.UpdateTaskStatus(accountID, b.TxID, Failed, taskTypeName, err.Error()))
	}

	log.Infof("Task %s successful for transaction:%v\n", taskTypeName, b.TxID.String())
	return b.TxManager.UpdateTaskStatus(accountID, b.TxID, Success, taskTypeName, "")
}
