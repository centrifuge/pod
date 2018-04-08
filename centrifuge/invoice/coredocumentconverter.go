package invoice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	invoicepb "github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/golang/protobuf/proto"
	"log"
	"github.com/golang/protobuf/ptypes/any"
)

func ConvertToCoreDocument(inv *invoicepb.InvoiceDocument) (coredoc coredocument.CoreDocument) {
	proto.Merge(&coredoc, inv.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData", err)
	}

	invoiceAny := any.Any{
		TypeUrl: invoicepb.InvoiceDocumentTypeUrl,
		Value: serializedInvoice,
	}

	coredoc.EmbeddedDocument = &invoiceAny
	return
}

func ConvertToInvoiceDocument(coredoc *coredocument.CoreDocument) (inv invoicepb.InvoiceDocument) {
	if coredoc.EmbeddedDocument.TypeUrl != invoicepb.InvoiceDocumentTypeUrl {
		log.Fatal("Trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	proto.Unmarshal(coredoc.EmbeddedDocument.Value, invoiceData)
	emptiedCoreDoc := coredocument.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredoc)
	emptiedCoreDoc.EmbeddedDocument = nil
	inv.Data = invoiceData
	inv.CoreDocument = &emptiedCoreDoc
	return
}