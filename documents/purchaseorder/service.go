package purchaseorder

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
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
	repo      documents.Repository
	queueSrv  queue.TaskQueuer
	txManager transactions.Manager
}

// DefaultService returns the default implementation of the service
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	txManager transactions.Manager,
) Service {
	return service{
		repo:      repo,
		queueSrv:  queueSrv,
		txManager: txManager,
		Service:   srv,
	}
}

// DeriveFromCoreDocument takes a core document model and returns a purchase order
func (s service) DeriveFromCoreDocumentModel(coreDocModel *documents.CoreDocumentModel) (documents.Model, error) {
	var model documents.Model = new(PurchaseOrder)
	err := model.UnpackCoreDocument(coreDocModel)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return model, nil
}

// validateAndPersist validates the document, and persists to DB
func (s service) validateAndPersist(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	self, err := contextutil.Self(ctx)
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
	err = s.repo.Create(self.ID[:], po.CoreDocumentModel.Document.CurrentVersion, po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return po, nil
}

// Create validates, persists, and anchors a purchase order
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, uuid.UUID, chan bool, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	po, err = s.validateAndPersist(ctx, nil, po, CreateValidator())
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	dm, err := po.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, self.ID, txID, dm.Document.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, nil, nil
	}
	return po, txID, done, nil
}

// Update validates, persists, and anchors a new version of purchase order
func (s service) Update(ctx context.Context, po documents.Model) (documents.Model, uuid.UUID, chan bool, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	dm, err := po.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	old, err := s.GetCurrentVersion(ctx, dm.Document.DocumentIdentifier)
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	po, err = s.validateAndPersist(ctx, old, po, UpdateValidator())
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, self.ID, txID, dm.Document.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}
	return po, txID, done, nil
}

// DeriveFromCreatePayload derives purchase order from create payload
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigAccountID
	}

	po := new(PurchaseOrder)
	err = po.InitPurchaseOrderInput(payload, idConf.ID.String())
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
	id, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentIdentifier, errors.New("failed to decode identifier: %v", err))
	}

	old, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	// load purchase order data
	po := new(PurchaseOrder)
	err = po.initPurchaseOrderFromData(payload.Data)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, errors.New("failed to load purchase order from data: %v", err))
	}

	// update core document
	oldDM, err := old.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigAccountID
	}

	collaborators := append([]string{idConf.ID.String()}, payload.Collaborators...)
	po.CoreDocumentModel, err = oldDM.PrepareNewVersion(collaborators)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPrepareCoreDocument, err)
	}

	return po, nil
}

// DerivePurchaseOrderData returns po data from the model
func (s service) DerivePurchaseOrderData(doc documents.Model) (*clientpopb.PurchaseOrderData, error) {
	po, ok := doc.(*PurchaseOrder)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return po.getClientData(), nil
}

// DerivePurchaseOrderResponse returns po response from the model
func (s service) DerivePurchaseOrderResponse(doc documents.Model) (*clientpopb.PurchaseOrderResponse, error) {
	dm, err := doc.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}
	cd := dm.Document
	collaborators := make([]string, len(cd.Collaborators))
	for i, c := range cd.Collaborators {
		cid, err := identity.ToCentID(c)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentCollaborator, err)
		}
		collaborators[i] = cid.String()
	}

	h := &clientpopb.ResponseHeader{
		DocumentId:    hexutil.Encode(cd.DocumentIdentifier),
		VersionId:     hexutil.Encode(cd.CurrentVersion),
		Collaborators: collaborators,
	}

	data, err := s.DerivePurchaseOrderData(doc)
	if err != nil {
		return nil, err
	}

	return &clientpopb.PurchaseOrderResponse{
		Header: h,
		Data:   data,
	}, nil
}
