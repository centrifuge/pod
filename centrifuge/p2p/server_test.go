package p2p

import (
	"testing"
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

var storageService = storage.LeveldbDataStore{Path:"/tmp/centrifuge_testing.leveldb"}

func TestP2PService(t *testing.T) {
	storage.SetStorage(&storageService)
	defer storageService.Close()

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

	doc, err := storageService.GetDocument(identifier)
	unmarshalledInv := invoice.ConvertToInvoiceDocument(doc)
	if unmarshalledInv.Data.Amount != inv.Data.Amount {
		t.Fatal("Invoice Amount doesn't match")
	}


	if !bytes.Equal(doc.DocumentIdentifier, identifier) {
		t.Fatal("Document Identifier doesn't match")
	}
}