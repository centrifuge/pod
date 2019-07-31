package pending

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	logging "github.com/ipfs/go-log"
)

var srvLog = logging.Logger("pending-service")

// Service provides an interface for functions common to all document types
type Service interface {
	// Commit validates, shares and anchors document
	Commit(ctx context.Context, model documents.Model) (jobs.JobID, error)
}

// service implements Service
type service struct {
	docSvc documents.Service
	// pendingRepo Repository
}

// DefaultService returns the default implementation of the service
func DefaultService(docSvc documents.Service) Service {
	return service{
		docSvc: docSvc,
	}
}

// Commit triggers validations, state change and anchor job
func (s service) Commit(ctx context.Context, model documents.Model) (jobs.JobID, error) {
	jobID, err := s.docSvc.Commit(ctx, model)
	if err != nil {
		return jobs.NilJobID(), err
	}

	// TODO Remove document from pending DB as soon as job is triggered successfully

	return jobID, nil
}
