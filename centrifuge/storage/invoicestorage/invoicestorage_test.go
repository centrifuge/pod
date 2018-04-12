// +build unit

package invoicestorage

import (
	"testing"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	invoicepb "github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/stretchr/testify/assert"
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

	invoice := invoicepb.InvoiceDocument{CoreDocument: &coredocument.CoreDocument{DocumentIdentifier:identifier}}
	service.PutDocument(&invoice)

	_, err := service.GetDocument(identifier)
	assert.Nil(t, err, "GetDocument should not return error")

	_, err = service.GetDocument(invalidIdentifier)
	assert.NotNil(t, err, "GetDocument should not return error")
}

