// +build integration unit testworld

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
		GrossAmount: []byte{0, 42, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func CreateInvoicePayload() *clientinvoicepb.InvoiceCreatePayload {
	return &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			Sender:      "0xed03Fa80291fF5DDC284DE6b51E716B130b05e20",
			Recipient:   "0xEA939D5C0494b072c51565b191eE59B5D34fbf79",
			Payee:       "0x087D8ca6A16E6ce8d9fF55672E551A2828Ab8e8C",
			GrossAmount: "42",
			Currency:    "EUR",
		},
		WriteAccess: []string{testingidentity.GenerateRandomDID().String()},
	}
}
