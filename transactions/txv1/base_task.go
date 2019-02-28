package txv1

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("transaction")

// BaseTask holds the required details and helper functions for tasks to update transactions.
// should be embedded into the task
type BaseTask struct {
	TxID transactions.TxID

	// state
	TxManager transactions.Manager
}

// ParseTransactionID parses txID.
func (b *BaseTask) ParseTransactionID(taskTypeName string, kwargs map[string]interface{}) error {
	txID, ok := kwargs[transactions.TxIDParam].(string)
	if !ok {
		return errors.New("missing transaction ID")
	}

	var err error
	b.TxID, err = transactions.FromString(txID)
	if err != nil {
		return errors.New("invalid transaction ID")
	}

	log.Infof("Task %s parsed for tx: %s\n", taskTypeName, b.TxID)
	return nil
}

// UpdateTransaction add a new log and updates the status of the transaction based on the error.
func (b *BaseTask) UpdateTransaction(accountID identity.DID, taskTypeName string, err error) error {
	if err == gocelery.ErrTaskRetryable {
		return err
	}

	// TODO this TaskStatus map update assumes that a single transaction has only one execution of a certain task type, which can be wrong, use the taskID or another unique identifier instead.
	if err != nil {
		log.Errorf("Task %s failed for transaction: %v with error: %s\n", taskTypeName, b.TxID.String(), err.Error())
		return errors.AppendError(err, b.TxManager.UpdateTaskStatus(accountID, b.TxID, transactions.Failed, taskTypeName, err.Error()))
	}

	log.Infof("Task %s successful for transaction:%v\n", taskTypeName, b.TxID.String())
	return b.TxManager.UpdateTaskStatus(accountID, b.TxID, transactions.Success, taskTypeName, "")
}
