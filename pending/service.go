package pending

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	logging "github.com/ipfs/go-log"
)

var srvLog = logging.Logger("pending-service")

// Service provides an interface for functions common to all document types
type Service interface {
	// Commit validates, shares and anchors document
	Commit(ctx context.Context, docID []byte) (jobs.JobID, error)
}

// service implements Service
type service struct {
	docSrv      documents.Service
	pendingRepo Repository
}

// DefaultService returns the default implementation of the service
func DefaultService(docSrv documents.Service, repo Repository) Service {
	return service{
		docSrv:      docSrv,
		pendingRepo: repo,
	}
}

// Commit triggers validations, state change and anchor job
func (s service) Commit(ctx context.Context, docID []byte) (jobs.JobID, error) {
	accID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return jobs.NilJobID(), contextutil.ErrDIDMissingFromContext
	}

	model, err := s.pendingRepo.Get(accID[:], docID)
	if err != nil {
		return jobs.NilJobID(), documents.ErrDocumentNotFound
	}

	jobID, err := s.docSrv.Commit(ctx, model)
	if err != nil {
		return jobs.NilJobID(), err
	}

	return jobID, s.pendingRepo.Delete(accID[:], docID)
}
