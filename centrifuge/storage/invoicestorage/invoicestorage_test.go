package invoicestorage

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"os"
)

var dbFileName = "/tmp/centrifuge_testing_invdoc.leveldb"
var storageDb storage.LeveldbDataStore

func TestMain(m *testing.M) {
	storageDb = storage.LeveldbDataStore{Path: dbFileName}
	storageDb.Open()
	defer storageDb.Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestStorageService(t *testing.T) {
	service := StorageService{}

	service.storage = &storageDb

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

