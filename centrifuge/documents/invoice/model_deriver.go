package invoice

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
)


// InvoiceModel specific deriver
type ModelDeriver interface {
	// Embedded ModelDeriver
	documents.ModelDeriver

	DeriveWithInvoiceInput(*InvoiceInput) (documents.Model, error)
}


type InvoiceInput struct {
	// TODO add parameters according to new client API
}
