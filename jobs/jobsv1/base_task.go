package jobsv1

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/gocelery"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("jobs")

// BaseTask holds the required details and helper functions for tasks to update jobs.
// should be embedded into the task
type BaseTask struct {
	JobID jobs.JobID

	// state
	JobManager jobs.Manager
}

// ParseJobID parses JobID.
func (b *BaseTask) ParseJobID(taskTypeName string, kwargs map[string]interface{}) error {
	jobID, ok := kwargs[jobs.JobIDParam].(string)
	if !ok {
		return errors.New("missing job ID")
	}

	var err error
	b.JobID, err = jobs.FromString(jobID)
	if err != nil {
		return errors.New("invalid job ID")
	}

	log.Infof("Task %s parsed for tx: %s\n", taskTypeName, b.JobID)
	return nil
}

// UpdateJob add a new log and updates the status of the job based on the error.
func (b *BaseTask) UpdateJob(accountID identity.DID, taskTypeName string, err error) error {
	return b.UpdateJobWithValue(accountID, taskTypeName, err, nil)
}

// UpdateJobWithValue add a new log and updates the status of the transaction based on the error and adds a value to the tx
func (b *BaseTask) UpdateJobWithValue(accountID identity.DID, taskTypeName string, err error, txValue *jobs.JobValue) error {
	if err == gocelery.ErrTaskRetryable {
		return err
	}

	// TODO this TaskStatus map update assumes that a single transaction has only one execution of a certain task type, which can be wrong, use the taskID or another unique identifier instead.
	if err != nil {
		log.Errorf("Task %s failed for job: %v with error: %s\n", taskTypeName, b.JobID.String(), err.Error())
		return errors.AppendError(err, b.JobManager.UpdateTaskStatus(accountID, b.JobID, jobs.Failed, taskTypeName, err.Error()))
	}

	log.Infof("Task %s successful for job:%v\n", taskTypeName, b.JobID.String())
	if txValue != nil {
		err = b.JobManager.UpdateJobWithValue(accountID, b.JobID, txValue.Key, txValue.Value)
		if err != nil {
			return err
		}
	}
	return b.JobManager.UpdateTaskStatus(accountID, b.JobID, jobs.Success, taskTypeName, "")
}
