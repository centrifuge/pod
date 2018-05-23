package invoice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"log"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/golang/protobuf/proto"
)

type Invoice struct {
	Document *invoicepb.InvoiceDocument
}

func NewInvoice(invDoc *invoicepb.InvoiceDocument) *Invoice {
	inv := &Invoice{invDoc}
	// IF salts have not been provided, let's generate them
	if invDoc.Salts == nil {
		invoiceSalts := invoicepb.InvoiceDataSalts{}
		proofs.FillSalts(&invoiceSalts)
		inv.Document.Salts = &invoiceSalts
	}
	return inv
}

func NewEmptyInvoice() *Invoice {
	invoiceSalts := invoicepb.InvoiceDataSalts{}
	proofs.FillSalts(&invoiceSalts)
	doc := invoicepb.InvoiceDocument{
		CoreDocument: &coredocumentpb.CoreDocument{},
		Data: &invoicepb.InvoiceData{},
		Salts: &invoiceSalts,
	}
	return &Invoice{&doc}
}

func NewInvoiceFromCoreDocument(coredocument *coredocument.CoreDocument) (inv *Invoice) {
	if coredocument.Document.EmbeddedData.TypeUrl != documenttypes.InvoiceDataTypeUrl ||
		coredocument.Document.EmbeddedDataSalts.TypeUrl != documenttypes.InvoiceSaltsTypeUrl {
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
	inv = NewEmptyInvoice()
	inv.Document.Data = invoiceData
	inv.Document.Salts = invoiceSalts
	inv.Document.CoreDocument = &emptiedCoreDoc
	return
}

func (inv *Invoice) CalculateMerkleRoot() {
	dtree := proofs.NewDocumentTree()
	dtree.FillTree(inv.Document.Data, inv.Document.Salts)
	inv.Document.CoreDocument.DocumentRoot = dtree.RootHash()
}

func (inv *Invoice) ConvertToCoreDocument() (coredocument coredocument.CoreDocument) {
	coredocpb := &coredocumentpb.CoreDocument{}
	proto.Merge(coredocpb, inv.Document.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Document.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData: %s", err)
	}

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value: serializedInvoice,
	}

	serializedSalts, err := proto.Marshal(inv.Document.Salts)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceSalts: %s", err)
	}

	invoiceSaltsAny := any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		Value: serializedSalts,
	}

	coredocpb.EmbeddedData = &invoiceAny
	coredocpb.EmbeddedDataSalts = &invoiceSaltsAny
	coredocument.Document = coredocpb
	return
}
