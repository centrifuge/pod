package invoice

import (
	"bytes"
	"encoding/hex"
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

	// DeriveInvoiceData returns the invoice data as client data
	DeriveInvoiceData(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// DeriveInvoiceResponse returns the invoice model in our standard client format
	DeriveInvoiceResponse(inv documents.Model) (*clientinvoicepb.InvoiceResponse, error)

	// GetVersion reads a document from the database
	GetVersion([]byte, []byte) (documents.Model, error)
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

	// we use currentIdentifier as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(coreDoc.CurrentIdentifier, inv)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	return nil
}

func (s service) GetVersion(identifier []byte, version []byte) (doc documents.Model, err error) {
	doc = new(InvoiceModel)
	err = s.repo.LoadByID(version, doc)
	if err != nil {
		return nil, err
	}

	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "not an invoice object")
	}

	if !bytes.Equal(inv.CoreDocument.DocumentIdentifier, identifier) {
		return nil, centerrors.New(code.DocumentInvalid, "version is not valid for this identifier")
	}
	return
}

// DeriveInvoiceResponse returns create response from invoice model
func (s service) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}
	collaborators := make([]string, len(inv.Collaborators))
	for i, c := range inv.Collaborators {
		collaborators[i] = c.String()
	}

	header := &clientinvoicepb.ResponseHeader{
		DocumentId:    hex.EncodeToString(inv.CoreDocument.DocumentIdentifier),
		VersionId:     hex.EncodeToString(inv.CoreDocument.CurrentIdentifier),
		Collaborators: collaborators,
	}

	data, err := s.DeriveInvoiceData(doc)
	if err != nil {
		return nil, err
	}

	return &clientinvoicepb.InvoiceResponse{
		Header: header,
		Data:   data,
	}, nil

}

// DeriveInvoiceData returns create response from invoice model
func (s service) DeriveInvoiceData(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
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
