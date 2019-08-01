package pending

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	logging "github.com/ipfs/go-log"
)

var srvLog = logging.Logger("pending-service")

// ErrPendingDocumentExists is a sentinel error used when document was created and tried to create a new one.
const ErrPendingDocumentExists = errors.Error("Pending document already created")

// ErrInProcessDocument is a sentinel error used when a document is not in pending state
const ErrInProcessDocument = errors.Error("document is not in pending state")

// Service provides an interface for functions common to all document types
type Service interface {

	// Update updates a pending document from the payload
	Update(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error)

	// Create creates a pending document from the payload
	Create(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error)

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

// Create creates either a new document or next version of an anchored document and stores the document.
// errors out if there an pending document created already
func (s service) Create(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error) {
	accID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, contextutil.ErrDIDMissingFromContext
	}

	if len(payload.DocumentID) > 0 {
		_, err := s.pendingRepo.Get(accID[:], payload.DocumentID)
		if err == nil {
			// found an existing pending document. error out
			return nil, ErrPendingDocumentExists
		}
	}

	doc, err := s.docSrv.Derive(ctx, payload)
	if err != nil {
		return nil, err
	}

	// we create one document per ID. hence, we use ID instead of current version
	// since its common to all document versions.
	return doc, s.pendingRepo.Create(accID[:], doc.ID(), doc)
}

// Update updates a pending document from the payload
func (s service) Update(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error) {
	accID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, contextutil.ErrDIDMissingFromContext
	}

	m, err := s.pendingRepo.Get(accID[:], payload.DocumentID)
	if err != nil {
		return nil, err
	}

	if m.GetStatus() != documents.Pending {
		return nil, ErrInProcessDocument
	}

	mp, ok := m.(documents.Patcher)
	if !ok {
		return nil, documents.ErrNotPatcher
	}

	doc, err := mp.Patch(payload)
	if err != nil {
		return nil, err
	}

	return doc, s.pendingRepo.Update(accID[:], doc.ID(), doc)
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
