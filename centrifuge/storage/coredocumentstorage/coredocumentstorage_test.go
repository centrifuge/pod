// +build unit

package coredocumentstorage

import (
	"testing"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"os"
)

var dbFileName = "/tmp/centrifuge_testing_coredoc.leveldb"
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
	coredoc := coredocumentpb.CoreDocument{DocumentIdentifier:identifier}
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

