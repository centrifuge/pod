// +build unit

package invoice


import (
	"testing"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)


func TestCoreDocumentConverter(t *testing.T) {
	identifier := []byte("1")
	invoiceData := invoicepb.InvoiceData{
		Amount: 100,
	}
	invoiceSalts := invoicepb.InvoiceDataSalts{}

	invoiceDoc := NewInvoiceDocument()
	invoiceDoc.CoreDocument =	&coredocumentpb.CoreDocument{
			DocumentIdentifier:identifier,
			}
	invoiceDoc.Data = &invoiceData
	invoiceDoc.Salts = &invoiceSalts

	serializedInvoice, err := proto.Marshal(&invoiceData)
	assert.Nil(t, err, "Could not serialize InvoiceData")

	serializedSalts, err := proto.Marshal(&invoiceSalts)
	assert.Nil(t, err, "Could not serialize InvoiceDataSalts")

	invoiceAny := any.Any{
		TypeUrl: invoicepb.InvoiceDataTypeUrl,
		Value: serializedInvoice,
	}
	invoiceSaltsAny := any.Any{
		TypeUrl: invoicepb.InvoiceSaltsTypeUrl,
		Value: serializedSalts,
	}
	coreDocument := coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData: &invoiceAny,
		EmbeddedDataSalts: &invoiceSaltsAny,
	}

	generatedCoreDocument := ConvertToCoreDocument(invoiceDoc)
	generatedCoreDocumentBytes, err := proto.Marshal(&generatedCoreDocument)
	assert.Nil(t, err, "Error marshaling generatedCoreDocument")

	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	assert.Nil(t, err, "Error marshaling coreDocument")
	assert.Equal(t, coreDocumentBytes, generatedCoreDocumentBytes,
		"Generated & converted documents are not identical")


	convertedInvoiceDoc := ConvertToInvoiceDocument(&generatedCoreDocument)
	convertedGeneratedInvoiceDoc := ConvertToInvoiceDocument(&generatedCoreDocument)
	invoiceBytes, err := proto.Marshal(invoiceDoc)
	assert.Nil(t, err, "Error marshaling invoiceDoc")

	convertedGeneratedInvoiceBytes, err := proto.Marshal(&convertedGeneratedInvoiceDoc)
	assert.Nil(t, err, "Error marshaling convertedGeneratedInvoiceDoc")

	convertedInvoiceBytes, err := proto.Marshal(&convertedInvoiceDoc)
	assert.Nil(t, err, "Error marshaling convertedGeneratedInvoiceDoc")

	assert.Equal(t, invoiceBytes, convertedGeneratedInvoiceBytes,
		"invoiceBytes and convertedGeneratedInvoiceBytes do not match")
	assert.Equal(t, invoiceBytes, convertedInvoiceBytes,
		"invoiceBytes and convertedInvoiceBytes do not match")

}

