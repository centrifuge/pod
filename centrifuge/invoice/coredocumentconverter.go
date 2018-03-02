package invoice

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/gogo/protobuf/proto"
	"log"
	"github.com/golang/protobuf/ptypes/any"
)

func ConvertToCoreDocument(inv *InvoiceDocument) (coredoc coredocument.CoreDocument) {
	proto.Merge(&coredoc, inv.CoreDocument)
	marshaledInvoiceData, err := proto.Marshal(inv.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData")
	}

	invoiceAny := any.Any{
		TypeUrl: "https://github.com/CentrifugeInc/go-centrifuge/blob/master/centrifuge/invoice/invoice.proto#invoice.InvoiceDocument",
		Value:   marshaledInvoiceData,
	}

	coredoc.EmbeddedDocument = &invoiceAny
	return
}

func ConvertToInvoiceDocument(coredoc *coredocument.CoreDocument) (inv InvoiceDocument) {
	if coredoc.EmbeddedDocument.TypeUrl != "https://github.com/CentrifugeInc/go-centrifuge/blob/master/centrifuge/invoice/invoice.proto#invoice.InvoiceDocument" {
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