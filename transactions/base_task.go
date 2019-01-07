package transactions

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
)

const (
	// TxIDParam maps transaction ID in the kwargs.
	TxIDParam = "transactionID"
)

// BaseTask holds the required details and helper functions for tasks to update transactions.
// should be embedded into the task
type BaseTask struct {
	TxID      uuid.UUID
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

	return nil
}

// UpdateTransaction add a new log and updates the status of the transaction based on the error.
func (b *BaseTask) UpdateTransaction(tenantID common.Address, name string, err error) error {
	tx, erri := b.TxService.GetTransaction(tenantID, b.TxID)
	if erri != nil {
		return errors.AppendError(err, erri)
	}

	if err == nil {
		tx.Status = Success
		tx.Logs = append(tx.Logs, NewLog(name, ""))
		return b.TxService.SaveTransaction(tx)
	}

	tx.Status = Failed
	tx.Logs = append(tx.Logs, NewLog(name, err.Error()))
	return errors.AppendError(err, b.TxService.SaveTransaction(tx))
}
