package invoice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
)

type Invoice struct {
	Document *invoicepb.InvoiceDocument
}

func NewInvoice(invDoc *invoicepb.InvoiceDocument) *Invoice {
	return &Invoice{invDoc}
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

func (inv *Invoice) CalculateMerkleRoot() {
	dtree := proofs.NewDocumentTree()
	dtree.FillTree(inv.Document.Data, inv.Document.Salts)
	inv.Document.CoreDocument.DocumentRoot = dtree.RootHash()
}
