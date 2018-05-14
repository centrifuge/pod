package invoice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"crypto/sha256"
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
	hashFunc := sha256.New()
	dtree.SetHashFunc(hashFunc)
	err := dtree.FillTree(doc.Data, doc.Salts)
	if err != nil {
		panic(err)
	}
	doc.CoreDocument.DocumentRoot = dtree.RootHash()
}
