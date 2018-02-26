package coredocumentstorage

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
	identifier := []byte("1")
	invalidIdentifier := []byte("2")
	coredoc := coredocument.CoreDocument{DocumentIdentifier:identifier}
	service.PutDocument(&coredoc)

	doc, err := service.GetDocument(identifier)
	if err != nil {
		t.Fatal("Error getting document")
	}

	if !bytes.Equal(doc.DocumentIdentifier, identifier) {
		t.Fatal("Id doesn't match")
	}

	_, err = service.GetDocument(invalidIdentifier)
	if err == nil {
		t.Fatal("Should return error when getting invalid document")
	}
}

