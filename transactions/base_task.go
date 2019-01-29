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

	// TxNextTask indicates if there is next task
	TxNextTask = "next task"
)

// BaseTask holds the required details and helper functions for tasks to update transactions.
// should be embedded into the task
type BaseTask struct {
	TxID uuid.UUID

	// TODO [TXManager] remove this update once TX Manager update is complete, i.e. Individual tasks must not be responsible for updating a transactions status
	Next bool

	// state
	TxManager Manager
}

// ParseTransactionID parses txID.
func (b *BaseTask) ParseTransactionID(kwargs map[string]interface{}) error {
	txID, ok := kwargs[TxIDParam].(string)
	if !ok {
		return errors.New("missing transaction ID")
	}

	var err error
	b.TxID, err = uuid.FromString(txID)
	if err != nil {
		return errors.New("invalid transaction ID")
	}

	if b.Next, ok = kwargs[TxNextTask].(bool); !ok {
		b.Next = false
	}

	log.Infof("Task %v has next task: %v\n", b.TxID.String(), b.Next)
	return nil
}

// UpdateTransaction add a new log and updates the status of the transaction based on the error.
func (b *BaseTask) UpdateTransaction(accountID identity.CentID, taskTypeName string, err error) error {
	if err == gocelery.ErrTaskRetryable {
		return err
	}

	if err != nil {
		log.Infof("Transaction failed: %v\n", b.TxID.String())
		return errors.AppendError(err, b.updateStatus(accountID, Failed, taskTypeName, err.Error()))
	}

	if b.Next {
		return b.updateStatus(accountID, Pending, taskTypeName, "")
	}

	log.Infof("Transaction successful:%v\n", b.TxID.String())
	return b.updateStatus(accountID, Success, taskTypeName, "")
}

func (b *BaseTask) updateStatus(accountID identity.CentID, status Status, taskTypeName, message string) error {
	tx, err := b.TxManager.GetTransaction(accountID, b.TxID)
	if err != nil {
		return err
	}

	// TODO [TXManager] remove this update once TX Manager update is complete, i.e. Individual tasks must not be responsible for updating a transactions status
	tx.Status = status
	// status particular to the task
	tx.TaskStatus[taskTypeName] = status
	tx.Logs = append(tx.Logs, NewLog(taskTypeName, message))
	return b.TxManager.SaveTransaction(tx)
}
