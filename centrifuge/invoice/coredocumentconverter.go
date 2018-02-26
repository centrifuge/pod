package invoice

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/gogo/protobuf/proto"
	"log"
	"github.com/golang/protobuf/ptypes/any"
	"bytes"
)

func ConvertToCoreDocument(inv *InvoiceDocument) (coredoc coredocument.CoreDocument) {
	proto.Merge(&coredoc, inv.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData")
	}

	invoiceAny := any.Any{
		TypeUrl: "http://github.com/CentrifugeInc/go-centrifuge/invoice/#"+proto.MessageName(inv.Data),
		Value: serializedInvoice,
	}

	coredoc.DocumentSchemaId = []byte(coredocument.InvoiceSchema)
	coredoc.EmbeddedDocument = &invoiceAny
	return
}

func ConvertToInvoiceDocument(coredoc *coredocument.CoreDocument) (inv InvoiceDocument) {
	if !bytes.Equal(coredoc.DocumentSchemaId, []byte(coredocument.InvoiceSchema)) {
		log.Fatal("Trying to convert document with incorrect schema")
	}

	invoiceData := &InvoiceData{}
	proto.Unmarshal(coredoc.EmbeddedDocument.Value, invoiceData)
	emptiedCoreDoc := coredocument.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredoc)
	emptiedCoreDoc.EmbeddedDocument = nil
	inv.Data = invoiceData
	inv.CoreDocument = &emptiedCoreDoc
	return
}