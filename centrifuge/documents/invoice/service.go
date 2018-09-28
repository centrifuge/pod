package invoice

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
)

type service struct {
}

//DeriveWithInvoiceInput can initialize the model with parameters provided from the rest-api call
func (s *service) DeriveWithInvoiceInput(invoiceInput *InvoiceInput) (documents.Model, error) {

	if invoiceInput == nil {
		return nil, fmt.Errorf("invoiceInput should not be nil")
	}

	invoiceModel := new(InvoiceModel)

	invoiceModel.InitInvoiceInput(invoiceInput)

	return invoiceModel, nil

}

//DeriveWithCoreDocument can initialize the model with a coredocument received.
//Example: received coreDocument form other p2p node could use DeriveWithCoreDocument
func (s *service) DeriveWithCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {

	var model documents.Model

	model = new(InvoiceModel)

	err := model.FromCoreDocument(cd)

	if err != nil {
		return nil, err
	}

	return model, nil
}
