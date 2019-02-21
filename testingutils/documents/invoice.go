// +build integration unit

package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
)

func CreateInvoiceData() invoicepb.InvoiceData {
	recipient := testingidentity.GenerateRandomDID()
	sender := testingidentity.GenerateRandomDID()
	payee := testingidentity.GenerateRandomDID()
	return invoicepb.InvoiceData{
		Recipient:   recipient[:],
		Sender:      sender[:],
		Payee:       payee[:],
		GrossAmount: 42,
	}
}

func CreateInvoicePayload() *clientinvoicepb.InvoiceCreatePayload {
	return &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			Sender:      "0xed03fa80291ff5ddc284de6b51e716b130b05e20",
			Recipient:   "0xea939d5c0494b072c51565b191ee59b5d34fbf79",
			Payee:       "0x087d8ca6a16e6ce8d9ff55672e551a2828ab8e8c",
			GrossAmount: 42,
			ExtraData:   "0x01020302010203",
			Currency:    "EUR",
		},
		Collaborators: []string{testingidentity.GenerateRandomDID().String()},
	}
}
