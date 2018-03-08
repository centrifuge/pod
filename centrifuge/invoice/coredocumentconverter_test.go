// +build unit

package invoice


import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"bytes"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)


func TestCoreDocumentConverter(t *testing.T) {
	identifier := []byte("1")
	invoiceData := InvoiceData{
		Amount: 100,
	}
	invoiceDoc := InvoiceDocument{
		CoreDocument: &coredocument.CoreDocument{
			DocumentIdentifier:identifier,
			},
		Data: &invoiceData,
	}
	serializedInvoice, err := proto.Marshal(&invoiceData)
	if err != nil {
		t.Fatalf("Could not serialize InvoiceData")
	}

	invoiceAny := any.Any{
		TypeUrl: "http://github.com/CentrifugeInc/go-centrifuge/invoice/#"+proto.MessageName(&invoiceData),
		Value: serializedInvoice,
	}

	coreDocument := coredocument.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedDocument: &invoiceAny,
		DocumentSchemaId: []byte(coredocument.InvoiceSchema),
	}

	generatedCoreDocument := ConvertToCoreDocument(&invoiceDoc)
	generatedCoreDocumentBytes, err := proto.Marshal(&generatedCoreDocument)
	if err != nil {
		t.Fatal("Error marshaling generatedCoreDocument")
	}
	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	if err != nil {
		t.Fatal("Error marshaling coreDocument")
	}
	if !bytes.Equal(coreDocumentBytes, generatedCoreDocumentBytes) {
		t.Fatal("Generated & converted documents are not identical")
	}


	convertedInvoiceDoc := ConvertToInvoiceDocument(&generatedCoreDocument)
	convertedGeneratedInvoiceDoc := ConvertToInvoiceDocument(&generatedCoreDocument)
	invoiceBytes, err := proto.Marshal(&invoiceDoc)
	if err != nil {
		t.Fatal("Error marshaling invoiceDoc")
	}


	convertedGeneratedInvoiceBytes, err := proto.Marshal(&convertedGeneratedInvoiceDoc)
	if err != nil {
		t.Fatal("Error marshaling convertedGeneratedInvoiceDoc")
	}

	convertedInvoiceBytes, err := proto.Marshal(&convertedInvoiceDoc)
	if err != nil {
		t.Fatal("Error marshaling convertedGeneratedInvoiceDoc")
	}

	if !bytes.Equal(invoiceBytes, convertedGeneratedInvoiceBytes) {
		t.Fatal("invoiceBytes and convertedGeneratedInvoiceBytes do not match")
	}

	if !bytes.Equal(invoiceBytes, convertedInvoiceBytes) {
		t.Fatal("invoiceBytes and convertedInvoiceBytes do not match")
	}

}

