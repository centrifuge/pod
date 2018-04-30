package invoice

import (
	coredocumentpb "github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	invoicepb "github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/golang/protobuf/proto"
	"log"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
)

func ConvertToCoreDocument(inv *Invoice) (coredocument coredocument.CoreDocument) {
	coredocpb := &coredocumentpb.CoreDocument{}
	proto.Merge(coredocpb, inv.Document.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Document.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData", err)
	}

	invoiceAny := any.Any{
		TypeUrl: invoicepb.InvoiceDataTypeUrl,
		Value: serializedInvoice,
	}

	serializedSalts, err := proto.Marshal(inv.Document.Salts)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceSalts: %s", err)
	}

	invoiceSaltsAny := any.Any{
		TypeUrl: invoicepb.InvoiceSaltsTypeUrl,
		Value: serializedSalts,
	}

	coredocpb.EmbeddedData = &invoiceAny
	coredocpb.EmbeddedDataSalts = &invoiceSaltsAny
	coredocument.Document = coredocpb
	return
}

func ConvertToInvoiceDocument(coredocument *coredocument.CoreDocument) (inv Invoice) {
	if coredocument.Document.EmbeddedData.TypeUrl != invoicepb.InvoiceDataTypeUrl ||
		coredocument.Document.EmbeddedDataSalts.TypeUrl != invoicepb.InvoiceSaltsTypeUrl {
		log.Fatal("Trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	proto.Unmarshal(coredocument.Document.EmbeddedData.Value, invoiceData)

	invoiceSalts := &invoicepb.InvoiceDataSalts{}
	proto.Unmarshal(coredocument.Document.EmbeddedDataSalts.Value, invoiceSalts)

	emptiedCoreDoc := coredocumentpb.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredocument.Document)
	emptiedCoreDoc.EmbeddedData = nil
	emptiedCoreDoc.EmbeddedDataSalts = nil
	inv = *NewEmptyInvoice()
	inv.Document.Data = invoiceData
	inv.Document.Salts = invoiceSalts
	inv.Document.CoreDocument = &emptiedCoreDoc
	return
}