package invoicestorage

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

func TestStorageService(t *testing.T) {
	service := StorageService{}

	// We should figure out a nicer way to inject this here.
	service.storage = &storage.LeveldbDataStore{Path:"/tmp/centrifuge_testing.leveldb"}
	service.storage.Open()
	defer service.storage.Close()

	invoice := coredocument.InvoiceDocument{DocumentIdentifier: []byte("1"), CoreDocument: &coredocument.CoreDocument{DocumentIdentifier:[]byte("1")}}
	service.PutDocument(&invoice)

	doc, err := service.GetDocument([]byte("1"))
	if err != nil {
		t.Fatal("Error getting document")
	}

	if !bytes.Equal(doc.DocumentIdentifier, []byte("1")) {
		t.Fatal("Id doesn't match")
	}
}

