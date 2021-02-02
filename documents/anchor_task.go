package documents

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

const (
	// DocumentIDParam maps to model ID in the kwargs
	DocumentIDParam = "documentID"

	// AccountIDParam maps to account ID in the kwargs
	AccountIDParam = "accountID"

	documentAnchorTaskName = "Document Anchoring"
)

var log = logging.Logger("anchor_task")

type documentAnchorTask struct {
	jobsv1.BaseTask

	id        []byte
	accountID identity.DID

	// state
	config        config.Service
	processor     AnchorProcessor
	modelGetFunc  func(tenantID, id []byte) (Document, error)
	modelSaveFunc func(tenantID, id []byte, model Document) error
}

// TaskTypeName returns the name of the task.
func (d *documentAnchorTask) TaskTypeName() string {
	return documentAnchorTaskName
}

// ParseKwargs parses the kwargs.
func (d *documentAnchorTask) ParseKwargs(kwargs map[string]interface{}) error {
	err := d.ParseJobID(d.TaskTypeName(), kwargs)
	if err != nil {
		return err
	}

	modelID, ok := kwargs[DocumentIDParam].(string)
	if !ok {
		return errors.New("missing model ID")
	}

	d.id, err = hexutil.Decode(modelID)
	if err != nil {
		return errors.New("invalid model ID")
	}

	accountID, ok := kwargs[AccountIDParam].(string)
	if !ok {
		return errors.New("missing account ID")
	}

	d.accountID, err = identity.NewDIDFromString(accountID)
	if err != nil {
		return errors.New("invalid cent ID")
	}
	return nil
}

// Copy returns a new task with state.
func (d *documentAnchorTask) Copy() (gocelery.CeleryTask, error) {
	return &documentAnchorTask{
		BaseTask:      jobsv1.BaseTask{JobManager: d.JobManager},
		config:        d.config,
		processor:     d.processor,
		modelGetFunc:  d.modelGetFunc,
		modelSaveFunc: d.modelSaveFunc,
	}, nil
}

// RunTask anchors the document.
func (d *documentAnchorTask) RunTask() (res interface{}, err error) {
	log.Infof("starting anchor task for transaction: %s\n", d.JobID)
	defer func() {
		err = d.UpdateJob(d.accountID, d.TaskTypeName(), err)
	}()

	tc, err := d.config.GetAccount(d.accountID[:])
	if err != nil {
		log.Error(err)
		return nil, errors.New("failed to get header: %v", err)
	}
	jobCtx := contextutil.WithJob(context.Background(), d.JobID)
	ctxh, err := contextutil.New(jobCtx, tc)
	if err != nil {
		return false, errors.New("failed to get context header: %v", err)
	}

	model, err := d.modelGetFunc(d.accountID[:], d.id)
	if err != nil {
		return false, errors.New("failed to get model: %v", err)
	}

	if _, err = AnchorDocument(ctxh, model, d.processor, func(id []byte, model Document) error {
		return d.modelSaveFunc(d.accountID[:], id, model)
	}, tc.GetPrecommitEnabled()); err != nil {
		return false, errors.New("failed to anchor document: %v", err)
	}

	return true, nil
}

// initDocumentAnchorTask enqueues a new document anchor task for a given combination of accountID/modelID/txID.
func initDocumentAnchorTask(jobMan jobs.Manager, tq queue.TaskQueuer, accountID identity.DID, modelID []byte, jobID jobs.JobID) (queue.TaskResult, error) {
	params := map[string]interface{}{
		jobs.JobIDParam: jobID.String(),
		DocumentIDParam: hexutil.Encode(modelID),
		AccountIDParam:  accountID.String(),
	}

	err := jobMan.UpdateTaskStatus(accountID, jobID, jobs.Pending, documentAnchorTaskName, "init")
	if err != nil {
		return nil, err
	}

	tr, err := tq.EnqueueJob(documentAnchorTaskName, params)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// CreateAnchorJob creates a job for anchoring a document using jobs manager
func CreateAnchorJob(parentCtx context.Context, jobsMan jobs.Manager, tq queue.TaskQueuer, self identity.DID, jobID jobs.JobID, documentID []byte) (jobs.JobID, chan error, error) {
	jobID, done, err := jobsMan.ExecuteWithinJob(contextutil.Copy(parentCtx), self, jobID, "anchor document", func(accountID identity.DID, jobID jobs.JobID, jobsMan jobs.Manager, errChan chan<- error) {
		tr, err := initDocumentAnchorTask(jobsMan, tq, accountID, documentID, jobID)
		if err != nil {
			errChan <- err
			return
		}
		_, err = tr.Get(jobsMan.GetDefaultTaskTimeout())
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	})
	return jobID, done, err
}
