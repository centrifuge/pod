package invoicestorage

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
)

func TestStorageService(t *testing.T) {
	service := StorageService{}

	// We should figure out a nicer way to inject this here.
	service.storage = &storage.LeveldbDataStore{Path:"/tmp/centrifuge_testing.leveldb"}
	service.storage.Open()
	defer service.storage.Close()
	identifier := []byte("1")
	invalidIdentifier := []byte("2")

	invoice := invoice.InvoiceDocument{CoreDocument: &coredocument.CoreDocument{DocumentIdentifier:identifier}}
	service.PutDocument(&invoice)

	doc, err := service.GetDocument(identifier)
	if err != nil {
		t.Fatal("Error getting document")
	}

	if !bytes.Equal(doc.CoreDocument.DocumentIdentifier, identifier) {
		t.Fatal("Id doesn't match")
	}
	_, err = service.GetDocument(invalidIdentifier)
	if err == nil {
		t.Fatal("Should return error when getting invalid document")
	}
}

