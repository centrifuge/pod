package invoice

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
)

// Service defines specific functions for invoice
type Service interface {
	documents.Service

	// DeriverFromPayload derives Invoice from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives invoice model from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceUpdatePayload) (documents.Model, error)

	// DeriveInvoiceData returns the invoice data as client data
	DeriveInvoiceData(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// DeriveInvoiceResponse returns the invoice model in our standard client format
	DeriveInvoiceResponse(inv documents.Model) (*clientinvoicepb.InvoiceResponse, error)
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo      documents.Repository
	queueSrv  queue.TaskQueuer
	txManager transactions.Manager
}

// DefaultService returns the default implementation of the service.
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

// DeriveFromCoreDocument unpacks the core document into a model
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model = new(Invoice)
	err := model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return model, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	id, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigAccountID
	}

	invoiceModel := new(Invoice)
	err = invoiceModel.InitInvoiceInput(payload, id.ID.String())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return invoiceModel, nil
}

// calculateDataRoot validates the document, calculates the data root, and persists to DB
func (s service) calculateDataRoot(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	inv, ok := new.(*Invoice)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// create data root, has to be done at the model level to access fields
	err = inv.CalculateDataRoot()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate the invoice
	err = validator.Validate(old, inv)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(self.ID[:], inv.CoreDocument.CurrentVersion, inv)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return inv, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, inv documents.Model) (documents.Model, uuid.UUID, chan bool, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	inv, err = s.calculateDataRoot(ctx, nil, inv, CreateValidator())
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	cd, err := inv.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, self.ID, txID, cd.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}
	return inv, txID, done, nil
}

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx context.Context, inv documents.Model) (documents.Model, uuid.UUID, chan bool, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	cd, err := inv.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	old, err := s.GetCurrentVersion(ctx, cd.DocumentIdentifier)
	if err != nil {
		return nil, uuid.Nil, nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	inv, err = s.calculateDataRoot(ctx, old, inv, UpdateValidator())
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, self.ID, txID, cd.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}
	return inv, txID, done, nil
}

// DeriveInvoiceResponse returns create response from invoice model
func (s service) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	cd, err := doc.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	collaborators := make([]string, len(cd.Collaborators))
	for i, c := range cd.Collaborators {
		cid, err := identity.ToCentID(c)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentCollaborator, err)
		}
		collaborators[i] = cid.String()
	}

	h := &clientinvoicepb.ResponseHeader{
		DocumentId:    hexutil.Encode(cd.DocumentIdentifier),
		VersionId:     hexutil.Encode(cd.CurrentVersion),
		Collaborators: collaborators,
	}

	data, err := s.DeriveInvoiceData(doc)
	if err != nil {
		return nil, err
	}

	return &clientinvoicepb.InvoiceResponse{
		Header: h,
		Data:   data,
	}, nil

}

// DeriveInvoiceData returns create response from invoice model
func (s service) DeriveInvoiceData(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	inv, ok := doc.(*Invoice)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return inv.getClientData(), nil
}

// DeriveFromUpdatePayload returns a new version of the old invoice identified by identifier in payload
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceUpdatePayload) (documents.Model, error) {
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

	// load invoice data
	inv := new(Invoice)
	err = inv.initInvoiceFromData(payload.Data)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, errors.New("failed to load invoice from data: %v", err))
	}

	// update core document
	oldCD, err := old.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigAccountID
	}

	collaborators := append([]string{idConf.ID.String()}, payload.Collaborators...)
	inv.CoreDocument, err = coredocument.PrepareNewVersion(*oldCD, collaborators)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPrepareCoreDocument, err)
	}

	return inv, nil
}
