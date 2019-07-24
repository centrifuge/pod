package purchaseorder

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

// service implements Service and handles all purchase order related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo       documents.Repository
	queueSrv   queue.TaskQueuer
	jobManager jobs.Manager
	anchorRepo anchors.AnchorRepository
}

// DefaultService returns the default implementation of the service
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

// DeriveFromCoreDocument takes a core document model and returns a purchase order
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Model, error) {
	po := new(PurchaseOrder)
	err := po.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return po, nil
}

// validateAndPersist validates the document, and persists to DB
func (s service) validateAndPersist(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	po, ok := new.(*PurchaseOrder)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// validate the invoice
	err = validator.Validate(old, po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(selfDID[:], po.CurrentVersion(), po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return po, nil
}

// Create validates, persists, and anchors a purchase order
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, jobs.JobID, chan error, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	po, err = s.validateAndPersist(ctx, nil, po, CreateValidator())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, selfDID, jobID, po.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, nil
	}
	return po, jobID, done, nil
}

// Update validates, persists, and anchors a new version of purchase order
func (s service) Update(ctx context.Context, new documents.Model) (documents.Model, jobs.JobID, chan error, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	old, err := s.GetCurrentVersion(ctx, new.ID())
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	new, err = s.validateAndPersist(ctx, old, new, UpdateValidator(s.anchorRepo))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, selfDID, jobID, new.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return new, jobID, done, nil
}

// CreateModel creates purchase order from the payload, validates, persists, and returns the purchase order.
func (s service) CreateModel(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	if payload.Data == nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentNil
	}

	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	po := new(PurchaseOrder)
	if err := po.unpackFromCreatePayload(did, payload); err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate po
	err = CreateValidator().Validate(nil, po)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(did[:], po.CurrentVersion(), po)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, po.CurrentVersion())
	return po, jobID, err
}

// UpdateModel updates the migrates the current purchase order to next version with data from the update payload
func (s service) UpdateModel(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	if payload.Data == nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentNil
	}

	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	old, err := s.GetCurrentVersion(ctx, payload.DocumentID)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	oldPO, ok := old.(*PurchaseOrder)
	if !ok {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("%v is not a purchase order", hexutil.Encode(payload.DocumentID)))
	}

	po := new(PurchaseOrder)
	err = po.unpackFromUpdatePayload(oldPO, payload)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	err = UpdateValidator(s.anchorRepo).Validate(old, po)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	err = s.repo.Create(did[:], po.CurrentVersion(), po)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, po.CurrentVersion())
	return po, jobID, err
}
