package invoicerepository

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
)

var invoiceRepository InvoiceRepository

func GetInvoiceRepository() InvoiceRepository {
	return invoiceRepository
}

type InvoiceRepository interface {
	GetKey(id []byte) []byte
	FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error)

	// CreateOrUpdate functions similar to a REST HTTP PUT where the document is either created or updated regardless if it existed before
	CreateOrUpdate(inv *invoicepb.InvoiceDocument) (err error)

	// Create will only create a document initially. If the same document (as identified by its DocumentIdentifier) exists
	// the Create method will error out
	Create(inv *invoicepb.InvoiceDocument) (err error)
}
