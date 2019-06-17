package generic

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo       documents.Repository
	queueSrv   queue.TaskQueuer
	jobManager jobs.Manager
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager,
) documents.Service {
	return service{
		repo:       repo,
		queueSrv:   queueSrv,
		jobManager: jobManager,
		Service:    srv,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an Generic Doc
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Model, error) {
	g := new(Generic)
	err := g.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return g, nil
}

// CreateModel creates generic from the payload, validates, persists, and returns the generic.
func (s service) CreateModel(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	g := new(Generic)
	if err := g.unpackFromCreatePayload(did, payload); err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(did[:], g.CurrentVersion(), g)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, g.CurrentVersion())
	return g, jobID, err
}

// UpdateModel updates the migrates the current entity to next version with data from the update payload
func (s service) UpdateModel(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	old, err := s.GetCurrentVersion(ctx, payload.DocumentID)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	oldGeneric, ok := old.(*Generic)
	if !ok {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("%v is not a Generic Document", hexutil.Encode(payload.DocumentID)))
	}

	g := new(Generic)
	err = g.unpackFromUpdatePayload(oldGeneric, payload)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	err = s.repo.Create(did[:], g.CurrentVersion(), g)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, g.CurrentVersion())
	return g, jobID, err
}
