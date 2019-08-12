package generic

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
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
	anchorRepo anchors.AnchorRepository
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager,
	anchorRepo anchors.AnchorRepository,
) documents.Service {
	return service{
		repo:       repo,
		queueSrv:   queueSrv,
		jobManager: jobManager,
		Service:    srv,
		anchorRepo: anchorRepo,
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

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx context.Context, new documents.Model) (documents.Model, jobs.JobID, chan error, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	old, err := s.GetCurrentVersion(ctx, new.ID())
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	// validate the document
	err = UpdateValidator(s.anchorRepo).Validate(old, new)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(selfDID[:], new.CurrentVersion(), new)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, selfDID, jobID, new.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return new, jobID, done, nil
}

// CreateModel creates generic from the payload, validates, persists, and returns the generic.
func (s service) CreateModel(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	g := new(Generic)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	if err := g.DeriveFromCreatePayload(payload); err != nil {
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
	err = g.unpackFromUpdatePayloadOld(oldGeneric, payload)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate the generic document
	err = UpdateValidator(s.anchorRepo).Validate(old, g)
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

// New returns a new uninitialised Generic document.
func (s service) New(_ string) (documents.Model, error) {
	return new(Generic), nil
}

// Validate takes care of document validation
func (s service) Validate(ctx context.Context, model documents.Model, old documents.Model) error {
	return nil
}
