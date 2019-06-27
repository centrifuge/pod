package purchaseorder

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for purchase order
type Service interface {
	documents.Service

	// DeriverFromPayload derives purchase order from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives purchase order from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error)

	// DerivePurchaseOrderData returns the purchase order data as client data
	DerivePurchaseOrderData(po documents.Model) (*clientpopb.PurchaseOrderData, error)

	// DerivePurchaseOrderResponse returns the purchase order in our standard client format
	DerivePurchaseOrderResponse(po documents.Model) (*clientpopb.PurchaseOrderResponse, error)
}

// service implements Service and handles all purchase order related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo           documents.Repository
	queueSrv       queue.TaskQueuer
	jobManager     jobs.Manager
	tokenRegFinder func() documents.TokenRegistry
	anchorRepo     anchors.AnchorRepository
}

// DefaultService returns the default implementation of the service
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager,
	tokenRegFinder func() documents.TokenRegistry,
	anchorRepo anchors.AnchorRepository,
) Service {
	return service{
		repo:           repo,
		queueSrv:       queueSrv,
		jobManager:     jobManager,
		Service:        srv,
		tokenRegFinder: tokenRegFinder,
		anchorRepo:     anchorRepo,
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
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
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
func (s service) Update(ctx context.Context, new documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
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

// DeriveFromCreatePayload derives purchase order from create payload
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	self, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigAccountID
	}

	po := new(PurchaseOrder)
	err = po.InitPurchaseOrderInput(payload, self)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return po, nil
}

// DeriveFromUpdatePayload derives purchase order from update payload
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	// get latest old version of the document
	id, err := hexutil.Decode(payload.DocumentId)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentIdentifier, errors.New("failed to decode identifier: %v", err))
	}

	old, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	cs, err := documents.FromClientCollaboratorAccess(payload.ReadAccess, payload.WriteAccess)
	if err != nil {
		return nil, err
	}

	attrs, err := documents.FromClientAttributes(payload.Attributes)
	if err != nil {
		return nil, err
	}

	// load purchase order data
	po := new(PurchaseOrder)
	err = po.PrepareNewVersion(old, payload.Data, cs, attrs)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, errors.New("failed to load purchase order from data: %v", err))
	}

	return po, nil
}

// DerivePurchaseOrderData returns po data from the model
func (s service) DerivePurchaseOrderData(doc documents.Model) (*clientpopb.PurchaseOrderData, error) {
	po, ok := doc.(*PurchaseOrder)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return po.getClientData()
}

// DerivePurchaseOrderResponse returns po response from the model
func (s service) DerivePurchaseOrderResponse(doc documents.Model) (*clientpopb.PurchaseOrderResponse, error) {
	data, err := s.DerivePurchaseOrderData(doc)
	if err != nil {
		return nil, err
	}

	h, err := documents.DeriveResponseHeader(s.tokenRegFinder(), doc)
	if err != nil {
		return nil, err
	}

	attrs, err := documents.ToClientAttributes(doc.GetAttributes())
	if err != nil {
		return nil, err
	}

	return &clientpopb.PurchaseOrderResponse{
		Header:     h,
		Data:       data,
		Attributes: attrs,
	}, nil
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
