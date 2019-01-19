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
	Next bool

	// state
	TxService Service
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
func (b *BaseTask) UpdateTransaction(accountID identity.CentID, name string, err error) error {
	if err == gocelery.ErrTaskRetryable {
		return err
	}

	if err != nil {
		log.Infof("Transaction failed: %v\n", b.TxID.String())
		return errors.AppendError(err, b.updateStatus(accountID, Failed, NewLog(name, err.Error())))
	}

	if b.Next {
		err = b.updateStatus(accountID, Pending, NewLog(name, ""))
		if err != nil {
			return err
		}

		return nil
	}

	log.Infof("Transaction successful:%v\n", b.TxID.String())
	return b.updateStatus(accountID, Success, NewLog(name, ""))
}

func (b *BaseTask) updateStatus(accountID identity.CentID, status Status, log Log) error {
	tx, err := b.TxService.GetTransaction(accountID, b.TxID)
	if err != nil {
		return err
	}

	tx.Status = status
	tx.Logs = append(tx.Logs, log)
	return b.TxService.SaveTransaction(tx)
}
