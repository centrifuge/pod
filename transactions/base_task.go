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

	// NextTaskParam maps next task name in the kwargs
	NextTaskParam = "nextTask"
)

// BaseTask holds the required details and helper functions for tasks to update transactions.
// should be embedded into the task
type BaseTask struct {
	TxID     uuid.UUID
	NextTask string

	// state
	TxService Service
	ChainTask func(task string, txID uuid.UUID)
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

	rn, ok := kwargs[NextTaskParam]
	if !ok {
		return nil
	}

	if b.NextTask, ok = rn.(string); !ok {
		return errors.New("failed to read next task: %v", rn)
	}

	if b.ChainTask == nil {
		return errors.New("chain task func is nil but next task exists")
	}

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

	if b.NextTask != "" {
		err = b.updateStatus(accountID, Pending, NewLog(name, ""))
		if err != nil {
			return err
		}

		b.ChainTask(b.NextTask, b.TxID)
		return nil
	}

	log.Infof("Transaction successful:%v\n", b.TxID.String())
	return errors.AppendError(err, b.updateStatus(accountID, Success, NewLog(name, "")))
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
