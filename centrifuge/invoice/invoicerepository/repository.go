package invoicerepository

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
)

var invoiceRepository InvoiceRepository

func GetInvoiceRepository() InvoiceRepository {
	return invoiceRepository
}

type InvoiceRepository interface {
	GetKey(id []byte) ([]byte)
	FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error)
	Store(inv *invoicepb.InvoiceDocument) (err error)
}
