package invoice

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func createInvoiceData() invoicepb.InvoiceData {
	return invoicepb.InvoiceData{
		Recipient:   tools.RandomSlice(identity.CentIDByteLength),
		Sender:      tools.RandomSlice(identity.CentIDByteLength),
		Payee:       tools.RandomSlice(identity.CentIDByteLength),
		GrossAmount: 42,
	}
}

func createInvoiceInput() InvoiceInput {
	return InvoiceInput{GrossAmount: 42}
}

func createCDWithEmbeddedInvoice(t *testing.T, invoiceData invoicepb.InvoiceData) *coredocumentpb.CoreDocument {
	identifier := []byte("1")
	invoiceSalts := invoicepb.InvoiceDataSalts{}

	serializedInvoice, err := proto.Marshal(&invoiceData)
	assert.Nil(t, err, "Could not serialize InvoiceData")

	serializedSalts, err := proto.Marshal(&invoiceSalts)
	assert.Nil(t, err, "Could not serialize InvoiceDataSalts")

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}
	invoiceSaltsAny := any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		Value:   serializedSalts,
	}
	coreDocument := &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData:       &invoiceAny,
		EmbeddedDataSalts:  &invoiceSaltsAny,
	}
	return coreDocument
}
