// +build unit

package jobsv1

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
)

func TestDocumentAnchorTask_updateTransaction(t *testing.T) {
	task := new(BaseTask)

	accountID := testingidentity.GenerateRandomDID()
	name := "some task"
	task.JobID = jobs.NewJobID()
	task.JobManager = NewManager(&mockConfig{}, NewRepository(ctx[storage.BootstrappedDB].(storage.Repository)))

	// missing transaction with nil error
	err := task.UpdateTransaction(accountID, name, nil)
	err = errors.GetErrs(err)[0]
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	// missing transaction with error
	err = task.UpdateTransaction(accountID, name, errors.New("anchor error"))
	err = errors.GetErrs(err)[1]
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	// no error and success
	tx := jobs.NewJob(accountID, "")
	assert.NoError(t, task.JobManager.(extendedManager).saveJob(tx))
	task.JobID = tx.ID
	assert.NoError(t, task.UpdateTransaction(accountID, name, nil))
	tx, err = task.JobManager.GetJob(accountID, task.JobID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, jobs.Pending)
	assert.Equal(t, tx.TaskStatus[name], jobs.Success)
	assert.Len(t, tx.Logs, 1)

	// failed task
	tx = jobs.NewJob(accountID, "")
	assert.NoError(t, task.JobManager.(extendedManager).saveJob(tx))
	task.JobID = tx.ID
	err = task.UpdateTransaction(accountID, name, errors.New("anchor error"))
	assert.EqualError(t, errors.GetErrs(err)[0], "anchor error")
	tx, err = task.JobManager.GetJob(accountID, task.JobID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, jobs.Pending)
	assert.Equal(t, tx.TaskStatus[name], jobs.Failed)
	assert.Len(t, tx.Logs, 1)
}
