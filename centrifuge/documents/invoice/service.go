package invoice

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
)

// Service is a interface for invoice specific functions for deriving a valid model
type Service interface {
	// Embedded documents.ModelDeriver
	documents.ModelDeriver

	// DeriverFromPayload derives InvoiceModel from clientPayload
	DeriveFromPayload(*clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// Create validates and persists invoice Model
	Create(inv documents.Model) error
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `centerrors` with proper error code
type service struct {
	repo documents.Repository
}

// DeriveFromPayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromPayload(invoiceInput *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error) {
	if invoiceInput == nil {
		return nil, centerrors.New(code.DocumentInvalid, "input is nil")
	}

	invoiceModel := new(InvoiceModel)
	invoiceModel.InitInvoiceInput(invoiceInput)
	return invoiceModel, nil
}

//DeriveFromCoreDocument can initialize the model with a core document received.
//Example: received CoreDoc form other p2p node could use DeriveWithCoreDocument
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model
	model = new(InvoiceModel)
	err := model.FromCoreDocument(cd)
	if err != nil {
		return nil, centerrors.New(code.Unknown, "")
	}

	return model, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(inv documents.Model) error {
	coreDoc, err := inv.CoreDocument()
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	// we use currentIdentifier as the id instead of document Identifier
	err = s.repo.Create(coreDoc.CurrentIdentifier, inv)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	return nil
}
