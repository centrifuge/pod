package p2p

import (
	"testing"
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"bytes"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
)

func init() {
	mockBootstrap()
}

func TestP2PService(t *testing.T) {

	identifier := []byte("1")
	inv := invoice.InvoiceDocument{
		CoreDocument: &coredocument.CoreDocument{DocumentIdentifier: identifier},
		Data: &invoice.InvoiceData{Amount: 100},
	}

	coredoc := invoice.ConvertToCoreDocument(&inv)
	req := P2PMessage{Document: &coredoc}
	rpc := P2PService{}
	res, err := rpc.Transmit(context.Background(), &req)


	if err != nil {
		t.Fatal("Received error")

	}

	if !bytes.Equal(res.Document.DocumentIdentifier, identifier) {
		t.Fatal("Incorrect identifier")
	}

	doc, err := cc.Node.GetCoreDocumentStorageService().GetDocument(identifier)
	unmarshalledInv := invoice.ConvertToInvoiceDocument(doc)
	if unmarshalledInv.Data.Amount != inv.Data.Amount {
		t.Fatal("Invoice Amount doesn't match")
	}


	if !bytes.Equal(doc.DocumentIdentifier, identifier) {
		t.Fatal("Document Identifier doesn't match")
	}

}

func mockBootstrap() {
	(&cc.MockCentNode{}).BootstrapDependencies()
}