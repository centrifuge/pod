package invoice

import (
	coredocumentpb "github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	invoicepb "github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/golang/protobuf/proto"
	"log"
	"github.com/golang/protobuf/ptypes/any"
)

func ConvertToCoreDocument(inv *invoicepb.InvoiceDocument) (coredoc coredocumentpb.CoreDocument) {
	proto.Merge(&coredoc, inv.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData", err)
	}

	invoiceAny := any.Any{
		TypeUrl: invoicepb.InvoiceDataTypeUrl,
		Value: serializedInvoice,
	}

	serializedSalts, err := proto.Marshal(inv.Salts)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceSalts: %s", err)
	}

	invoiceSaltsAny := any.Any{
		TypeUrl: invoicepb.InvoiceSaltsTypeUrl,
		Value: serializedSalts,
	}

	coredoc.EmbeddedData = &invoiceAny
	coredoc.EmbeddedDataSalts = &invoiceSaltsAny
	return
}

func ConvertToInvoiceDocument(coredoc *coredocumentpb.CoreDocument) (inv invoicepb.InvoiceDocument) {
	if coredoc.EmbeddedData.TypeUrl != invoicepb.InvoiceDataTypeUrl ||
		coredoc.EmbeddedDataSalts.TypeUrl != invoicepb.InvoiceSaltsTypeUrl {
		log.Fatal("Trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	proto.Unmarshal(coredoc.EmbeddedData.Value, invoiceData)

	invoiceSalts := &invoicepb.InvoiceDataSalts{}
	proto.Unmarshal(coredoc.EmbeddedDataSalts.Value, invoiceSalts)

	emptiedCoreDoc := coredocumentpb.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredoc)
	emptiedCoreDoc.EmbeddedData = nil
	emptiedCoreDoc.EmbeddedDataSalts = nil
	inv.Data = invoiceData
	inv.Salts = invoiceSalts
	inv.CoreDocument = &emptiedCoreDoc
	return
}