// +build unit

package txv1

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/assert"
)

func TestDocumentAnchorTask_updateTransaction(t *testing.T) {
	task := new(BaseTask)

	accountID := identity.RandomCentID()
	name := "some task"
	task.TxID = transactions.NewTxID()
	task.TxManager = NewManager(&mockConfig{}, NewRepository(ctx[storage.BootstrappedDB].(storage.Repository)))

	// missing transaction with nil error
	err := task.UpdateTransaction(accountID, name, nil)
	err = errors.GetErrs(err)[0]
	assert.True(t, errors.IsOfType(transactions.ErrTransactionMissing, err))

	// missing transaction with error
	err = task.UpdateTransaction(accountID, name, errors.New("anchor error"))
	err = errors.GetErrs(err)[1]
	assert.True(t, errors.IsOfType(transactions.ErrTransactionMissing, err))

	// no error and success
	tx := transactions.NewTransaction(accountID, "")
	assert.NoError(t, task.TxManager.(extendedManager).saveTransaction(tx))
	task.TxID = tx.ID
	assert.NoError(t, task.UpdateTransaction(accountID, name, nil))
	tx, err = task.TxManager.GetTransaction(accountID, task.TxID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, transactions.Pending)
	assert.Equal(t, tx.TaskStatus[name], transactions.Success)
	assert.Len(t, tx.Logs, 1)

	// failed task
	tx = transactions.NewTransaction(accountID, "")
	assert.NoError(t, task.TxManager.(extendedManager).saveTransaction(tx))
	task.TxID = tx.ID
	err = task.UpdateTransaction(accountID, name, errors.New("anchor error"))
	assert.EqualError(t, errors.GetErrs(err)[0], "anchor error")
	tx, err = task.TxManager.GetTransaction(accountID, task.TxID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, transactions.Pending)
	assert.Equal(t, tx.TaskStatus[name], transactions.Failed)
	assert.Len(t, tx.Logs, 1)
}
