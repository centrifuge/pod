package pending

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
)

// Service provides an interface for functions common to all document types
type Service interface {
	// Commit validates, shares and anchors document
	Commit(ctx context.Context, model documents.Model) (jobs.JobID, error)
}

// service implements Service
type service struct {
	docSvc     documents.Service
	docRepo    documents.Repository
	queueSrv   queue.TaskQueuer
	jobManager jobs.Manager
	// pendingRepo Repository
}

// DefaultService returns the default implementation of the service
func DefaultService(
	docSvc documents.Service,
	docRepo documents.Repository,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager) Service {
	return service{
		docSvc:     docSvc,
		docRepo:    docRepo,
		queueSrv:   queueSrv,
		jobManager: jobManager,
	}
}

// Commit triggers validations, state change and anchor job
func (s service) Commit(ctx context.Context, model documents.Model) (jobs.JobID, error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	if err := s.docSvc.Validate(ctx, model); err != nil {
		return jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentValidation, err)
	}

	if err := model.SetStatus(documents.Committing); err != nil {
		return jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	err = s.docRepo.Create(did[:], model.CurrentVersion(), model)
	if err != nil {
		return jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, model.CurrentVersion())
	if err != nil {
		return jobs.NilJobID(), err
	}

	// TODO Remove document from pending DB as soon as job is triggered successfully

	return jobID, nil
}
