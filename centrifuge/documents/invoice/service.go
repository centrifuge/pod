package invoice

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
)

type Service struct {

}

//DeriveWithInvoiceInput can initialize the model with parameters provided from the api call
func (s *Service) DeriveWithInvoiceInput(*InvoiceInput) (documents.Model, error) {
	panic("implement me")
}

//DeriveWithCoreDocument can intialize the model with a coredocument recieved for example from a
func (s *Service) DeriveWithCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	panic("implement me")
}

