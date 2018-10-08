package invoice

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
)

// Service defines specific functions for invoice
// TODO(ved): see if this interface can be used across the documents
type Service interface {
	documents.ModelDeriver

	// DeriverFromPayload derives InvoiceModel from clientPayload
	DeriveFromCreatePayload(*clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// Create validates and persists invoice Model and returns a Updated model
	Create(ctx context.Context, inv documents.Model) (documents.Model, error)

	// DeriveCreateResponse returns the invoice data as client data
	DeriveCreateResponse(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// SaveState updates the model in DB
	SaveState(inv documents.Model) error
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `centerrors` with proper error code
type service struct {
	repo             documents.Repository
	coreDocProcessor coredocumentprocessor.Processor
}

// DefaultService returns the default implementation of the service
func DefaultService(repo documents.Repository, processor coredocumentprocessor.Processor) Service {
	return &service{repo: repo, coreDocProcessor: processor}
}

// DeriveFromCoreDocument unpacks the core document into a model
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model = new(InvoiceModel)
	err := model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return model, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(invoiceInput *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error) {
	if invoiceInput == nil {
		return nil, centerrors.New(code.DocumentInvalid, "input is nil")
	}

	invoiceModel := new(InvoiceModel)
	err := invoiceModel.InitInvoiceInput(invoiceInput)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return invoiceModel, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, model documents.Model) (documents.Model, error) {
	// Validate the model
	fv := fieldValidator()
	err := fv.Validate(nil, model)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	// create data root
	inv := model.(*InvoiceModel)
	err = inv.calculateDataRoot()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	// we use currentIdentifier as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(inv.CoreDocument.CurrentIdentifier, inv)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	saveState := func(coreDoc *coredocumentpb.CoreDocument) error {
		err := inv.UnpackCoreDocument(coreDoc)
		if err != nil {
			return err
		}

		return s.SaveState(inv)
	}

	coreDoc, err := inv.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	err = s.coreDocProcessor.Anchor(ctx, coreDoc, inv.Collaborators, saveState)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	coreDoc, err = inv.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	for _, id := range inv.Collaborators {
		err := s.coreDocProcessor.Send(ctx, coreDoc, id)
		if err != nil {
			log.Infof("failed to send anchored document: %v\n", err)
		}
	}

	return inv, nil
}

// DeriveCreateResponse returns create response from invoice model
func (s service) DeriveCreateResponse(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	data := inv.getClientData()
	return data, nil
}

// SaveState updates the model on DB
// This will disappear once we have common DB for every document
func (s service) SaveState(doc documents.Model) error {
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	if inv.CoreDocument == nil {
		return centerrors.New(code.DocumentInvalid, "core document missing")
	}

	err := s.repo.Update(inv.CoreDocument.CurrentIdentifier, inv)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	return nil
}
