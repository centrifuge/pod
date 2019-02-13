// +build integration unit

package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/utils"
)

func CreateInvoiceData() invoicepb.InvoiceData {
	return invoicepb.InvoiceData{
		Recipient:   utils.RandomSlice(identity.CentIDLength),
		Sender:      utils.RandomSlice(identity.CentIDLength),
		Payee:       utils.RandomSlice(identity.CentIDLength),
		GrossAmount: 42,
	}
}

func CreateInvoicePayload() *clientinvoicepb.InvoiceCreatePayload {
	return &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			Sender:      "0x010101010101",
			Recipient:   "0x010203040506",
			Payee:       "0x010203020406",
			GrossAmount: 42,
			ExtraData:   "0x01020302010203",
			Currency:    "EUR",
		},
		Collaborators: []string{"0x010101010101"},
	}
}
