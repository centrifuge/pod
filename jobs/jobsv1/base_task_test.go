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
	err := task.UpdateJob(accountID, name, nil)
	err = errors.GetErrs(err)[0]
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	// missing transaction with error
	err = task.UpdateJob(accountID, name, errors.New("anchor error"))
	err = errors.GetErrs(err)[1]
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	// no error and success
	job := jobs.NewJob(accountID, "")
	assert.NoError(t, task.JobManager.(extendedManager).saveJob(job))
	task.JobID = job.ID
	assert.NoError(t, task.UpdateJob(accountID, name, nil))
	job, err = task.JobManager.GetJob(accountID, task.JobID)
	assert.NoError(t, err)
	assert.Equal(t, job.Status, jobs.Pending)
	assert.Equal(t, job.TaskStatus[name], jobs.Success)
	assert.Len(t, job.Logs, 1)

	// failed task
	job = jobs.NewJob(accountID, "")
	assert.NoError(t, task.JobManager.(extendedManager).saveJob(job))
	task.JobID = job.ID
	err = task.UpdateJob(accountID, name, errors.New("anchor error"))
	assert.EqualError(t, errors.GetErrs(err)[0], "anchor error")
	job, err = task.JobManager.GetJob(accountID, task.JobID)
	assert.NoError(t, err)
	assert.Equal(t, job.Status, jobs.Pending)
	assert.Equal(t, job.TaskStatus[name], jobs.Failed)
	assert.Len(t, job.Logs, 1)
}
