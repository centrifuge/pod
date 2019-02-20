// +build unit

package transactions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestDocumentAnchorTask_updateTransaction(t *testing.T) {
	task := new(BaseTask)

	accountID := testingidentity.GenerateRandomDID()
	name := "some task"
	task.TxID = uuid.Must(uuid.NewV4())
	task.TxManager = NewManager(&mockConfig{}, NewRepository(ctx[storage.BootstrappedDB].(storage.Repository)))

	// missing transaction with nil error
	err := task.UpdateTransaction(accountID, name, nil)
	err = errors.GetErrs(err)[0]
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	// missing transaction with error
	err = task.UpdateTransaction(accountID, name, errors.New("anchor error"))
	err = errors.GetErrs(err)[1]
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	// no error and success
	tx := newTransaction(accountID, "")
	assert.NoError(t, task.TxManager.(extendedManager).saveTransaction(tx))
	task.TxID = tx.ID
	assert.NoError(t, task.UpdateTransaction(accountID, name, nil))
	tx, err = task.TxManager.GetTransaction(accountID, task.TxID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, Pending)
	assert.Equal(t, tx.TaskStatus[name], Success)
	assert.Len(t, tx.Logs, 1)

	// failed task
	tx = newTransaction(accountID, "")
	assert.NoError(t, task.TxManager.(extendedManager).saveTransaction(tx))
	task.TxID = tx.ID
	err = task.UpdateTransaction(accountID, name, errors.New("anchor error"))
	assert.EqualError(t, errors.GetErrs(err)[0], "anchor error")
	tx, err = task.TxManager.GetTransaction(accountID, task.TxID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, Pending)
	assert.Equal(t, tx.TaskStatus[name], Failed)
	assert.Len(t, tx.Logs, 1)
}
