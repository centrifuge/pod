package server

import (
	"testing"
	"context"
	pb "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"bytes"
	"github.com/spf13/viper"
)

func TestCoreDocumentService(t *testing.T) {
	viper.SetDefault("storage.path", "/tmp/centrifuge_storage_testing.db")

	s := newCentrifugeNodeService()
	db := storage.GetStorage()
	doc := pb.InvoiceDocument{
		DocumentIdentifier:[]byte("1"),
	}
	s.SendInvoiceDocument(context.Background(), &pb.SendInvoiceEnvelope{[][]byte{}, &doc})

	stored_doc := db.GetInvoiceDocument([]byte("1"))

	if stored_doc == nil {
		t.Fatal("Document not found in DB")
	}

	if !bytes.Equal(stored_doc.DocumentIdentifier, doc.DocumentIdentifier) {
		t.Fatal("DocumentIdentifier does not match")
	}
}