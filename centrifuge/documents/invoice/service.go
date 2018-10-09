package invoice

import (
	"bytes"
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
)

// Service defines specific functions for invoice
// TODO(ved): see if this interface can be used across the documents
type Service interface {
	documents.ModelDeriver

	// DeriverFromPayload derives InvoiceModel from clientPayload
	DeriveFromCreatePayload(*clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// Create validates and persists invoice Model and returns a Updated model
	Create(ctx context.Context, inv documents.Model) (documents.Model, error)

	// DeriveInvoiceData returns the invoice data as client data
	DeriveInvoiceData(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// DeriveInvoiceResponse returns the invoice model in our standard client format
	DeriveInvoiceResponse(inv documents.Model) (*clientinvoicepb.InvoiceResponse, error)

	// GetLastVersion reads a document from the database
	GetLastVersion(identifier []byte) (documents.Model, error)

	// GetVersion reads a document from the database
	GetVersion(identifier []byte, version []byte) (documents.Model, error)

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

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(inv.CoreDocument.CurrentVersion, inv)
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

	err = s.coreDocProcessor.Anchor(ctx, coreDoc, saveState)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	coreDoc, err = inv.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	for _, id := range coreDoc.Collaborators {
		cid, err := identity.ToCentID(id)
		if err != nil {
			return nil, centerrors.New(code.Unknown, err.Error())
		}
		err = s.coreDocProcessor.Send(ctx, coreDoc, cid)
		if err != nil {
			log.Infof("failed to send anchored document: %v\n", err)
		}
	}

	return inv, nil
}

// GetVersion returns an invoice for a given version
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

// GetLastVersion returns the last known version of an invoice
func (s service) GetLastVersion(identifier []byte) (doc documents.Model, err error) {
	doc, err = s.GetVersion(identifier, identifier)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found")
	}
	inv := doc.(*InvoiceModel)
	nextVersion := inv.CoreDocument.NextVersion
	for nextVersion != nil {
		doc, err = s.GetVersion(identifier, nextVersion)
		if err != nil {
			return inv, nil
		} else {
			inv = doc.(*InvoiceModel)
			nextVersion = inv.CoreDocument.NextVersion
		}
	}
	return inv, nil
}

// DeriveInvoiceResponse returns create response from invoice model
func (s service) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	cd, err := doc.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}
	collaborators := make([]string, len(cd.Collaborators))
	for i, c := range cd.Collaborators {
		cid, err := identity.ToCentID(c)
		if err != nil {
			return nil, centerrors.New(code.Unknown, err.Error())
		}
		collaborators[i] = cid.String()
	}

	header := &clientinvoicepb.ResponseHeader{
		DocumentId:    hexutil.Encode(inv.CoreDocument.DocumentIdentifier),
		VersionId:     hexutil.Encode(inv.CoreDocument.CurrentVersion),
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

	err := s.repo.Update(inv.CoreDocument.CurrentVersion, inv)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	return nil
}
