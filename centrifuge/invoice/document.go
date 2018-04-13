package invoice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
)

func NewInvoiceDocument () *invoicepb.InvoiceDocument {
	invoiceSalts := invoicepb.InvoiceDataSalts{}
	proofs.FillSalts(&invoiceSalts)
	doc := invoicepb.InvoiceDocument{
		CoreDocument: &coredocumentpb.CoreDocument{},
		Data: &invoicepb.InvoiceData{},
		Salts: &invoiceSalts,
	}
	return &doc
}

func CalculateMerkleRoot (doc *invoicepb.InvoiceDocument) {
	dtree := proofs.NewDocumentTree()
	dtree.FillTree(doc.Data, doc.Salts)
	doc.CoreDocument.DocumentRoot = dtree.RootHash()
}
