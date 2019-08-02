package pending

import (
	"bytes"
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

// Service provides an interface for functions common to all document types
type Service interface {
	// Get returns the document associated with docID and Status.
	Get(ctx context.Context, docID []byte, status documents.Status) (documents.Model, error)

	// GetVersion returns the document associated with docID and versionID.
	GetVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error)

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

// Get returns the document associated with docID
// If status is pending, we return the pending document from pending repo.
// else, we defer Get to document service.
func (s service) Get(ctx context.Context, docID []byte, status documents.Status) (documents.Model, error) {
	if status != documents.Pending {
		return s.docSrv.GetCurrentVersion(ctx, docID)
	}

	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, contextutil.ErrDIDMissingFromContext
	}

	return s.pendingRepo.Get(did[:], docID)
}

// GetVersion return the specific version of the document
// We try to fetch the version from the document service, if found return
// else look in pending repo for specific version.
func (s service) GetVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	doc, err := s.docSrv.GetVersion(ctx, docID, versionID)
	if err == nil {
		return doc, nil
	}

	accID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, contextutil.ErrDIDMissingFromContext
	}

	doc, err = s.pendingRepo.Get(accID[:], docID)
	if err != nil || !bytes.Equal(versionID, doc.CurrentVersion()) {
		return nil, documents.ErrDocumentNotFound
	}

	return doc, nil
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

	mp, ok := m.(documents.Patcher)
	if !ok {
		return nil, documents.ErrNotPatcher
	}

	err = mp.Patch(payload)
	if err != nil {
		return nil, err
	}
	doc := mp.(documents.Model)
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
