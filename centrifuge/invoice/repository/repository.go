package repository

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
)

type InvoiceRepository interface {
	GetKey(id []byte) ([]byte)
	FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error)
	Store(inv *invoicepb.InvoiceDocument) (err error)
}
