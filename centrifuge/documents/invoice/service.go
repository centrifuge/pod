package invoice

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
)

// Service defines specific functions for invoice
// TODO(ved): see if this interface can be used across the documents
type Service interface {
	documents.ModelDeriver

	// DeriverFromPayload derives InvoiceModel from clientPayload
	DeriveFromCreatePayload(*clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// Create validates and persists invoice Model
	Create(inv documents.Model) error

	// DeriveCreateResponse returns the invoice data as client data
	DeriveCreateResponse(inv documents.Model) (*clientinvoicepb.InvoiceData, error)
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `centerrors` with proper error code
type service struct {
	repo documents.Repository
}

// DefaultService returns the default implementation of the service
func DefaultService(repo documents.Repository) Service {
	return &service{repo: repo}
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
	invoiceModel.InitInvoiceInput(invoiceInput)
	return invoiceModel, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(inv documents.Model) error {
	coreDoc, err := inv.PackCoreDocument()
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(coreDoc.CurrentVersion, inv)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	return nil
}

// DeriveCreateResponse returns create response from invoice model
func (s service) DeriveCreateResponse(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	data, err := inv.getClientData()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return data, nil
}
