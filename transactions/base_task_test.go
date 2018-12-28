// +build unit

package transactions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestDocumentAnchorTask_updateTransaction(t *testing.T) {
	task := new(BaseTask)

	tenantID := common.DummyIdentity
	name := "some task"
	task.TxID = uuid.Must(uuid.NewV4())
	task.TxRepository = NewRepository(ctx[storage.BootstrappedDB].(storage.Repository))

	// missing transaction with nil error
	err := task.UpdateTransaction(tenantID, name, nil)
	err = errors.GetErrs(err)[0]
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	// missing transaction with error
	err = task.UpdateTransaction(tenantID, name, errors.New("anchor error"))
	err = errors.GetErrs(err)[1]
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	// no error and success
	tx := NewTransaction(tenantID, "")
	assert.NoError(t, task.TxRepository.Save(tx))
	task.TxID = tx.ID
	assert.NoError(t, task.UpdateTransaction(tenantID, name, nil))
	tx, err = task.TxRepository.Get(tenantID, task.TxID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, Success)
	assert.Len(t, tx.Logs, 1)

	// failed task
	tx = NewTransaction(tenantID, "")
	assert.NoError(t, task.TxRepository.Save(tx))
	task.TxID = tx.ID
	err = task.UpdateTransaction(tenantID, name, errors.New("anchor error"))
	assert.EqualError(t, errors.GetErrs(err)[0], "anchor error")
	tx, err = task.TxRepository.Get(tenantID, task.TxID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, Failed)
	assert.Len(t, tx.Logs, 1)
}
