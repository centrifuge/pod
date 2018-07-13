// +build unit

package invoice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInvoiceCoreDocumentConverter(t *testing.T) {
	identifier := []byte("1")
	invoiceData := invoicepb.InvoiceData{
		GrossAmount: 100,
	}
	invoiceSalts := invoicepb.InvoiceDataSalts{}

	invoiceDoc := NewEmptyInvoice()
	invoiceDoc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
	}
	invoiceDoc.Document.Data = &invoiceData
	invoiceDoc.Document.Salts = &invoiceSalts

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
	coreDocument := coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData:       &invoiceAny,
		EmbeddedDataSalts:  &invoiceSaltsAny,
	}

	generatedCoreDocument := invoiceDoc.ConvertToCoreDocument()
	generatedCoreDocumentBytes, err := proto.Marshal(generatedCoreDocument)
	assert.Nil(t, err, "Error marshaling generatedCoreDocument")

	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	assert.Nil(t, err, "Error marshaling coreDocument")
	assert.Equal(t, coreDocumentBytes, generatedCoreDocumentBytes,
		"Generated & converted documents are not identical")

	convertedInvoiceDoc, err := NewInvoiceFromCoreDocument(generatedCoreDocument)
	convertedGeneratedInvoiceDoc, err := NewInvoiceFromCoreDocument(generatedCoreDocument)
	invoiceBytes, err := proto.Marshal(invoiceDoc.Document)
	assert.Nil(t, err, "Error marshaling invoiceDoc")

	convertedGeneratedInvoiceBytes, err := proto.Marshal(convertedGeneratedInvoiceDoc.Document)
	assert.Nil(t, err, "Error marshaling convertedGeneratedInvoiceDoc")

	convertedInvoiceBytes, err := proto.Marshal(convertedInvoiceDoc.Document)
	assert.Nil(t, err, "Error marshaling convertedGeneratedInvoiceDoc")

	assert.Equal(t, invoiceBytes, convertedGeneratedInvoiceBytes,
		"invoiceBytes and convertedGeneratedInvoiceBytes do not match")
	assert.Equal(t, invoiceBytes, convertedInvoiceBytes,
		"invoiceBytes and convertedInvoiceBytes do not match")

}
