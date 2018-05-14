// +build unit

package invoicerepository

import (
	"testing"
	"os"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
)

var dbFileName = "/tmp/centrifuge_testing_invdoc.leveldb"

func TestMain(m *testing.M) {
	defer Bootstrap().Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestStorageService(t *testing.T) {
	identifier := []byte("1")
	invalidIdentifier := []byte("2")

	invoice := invoicepb.InvoiceDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier:identifier}}
	repo := GetInvoiceRepository()
	err := repo.Store(&invoice)
	assert.Nil(t, err, "Store should not return error")

	inv, err := repo.FindById(identifier)
	assert.Nil(t, err, "FindById should not return error")
	assert.Equal(t, invoice.CoreDocument.DocumentIdentifier, inv.CoreDocument.DocumentIdentifier, "Invoice DocumentIdentifier should be equal")

	inv, err = repo.FindById(invalidIdentifier)
	assert.NotNil(t, err, "FindById should not return error")
	assert.Nil(t, inv, "Invoice should be NIL")
}

func Bootstrap() (*leveldb.DB) {
	levelDB := storage.NewLeveldbStorage(dbFileName)

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	NewLevelDBInvoiceRepository(&LevelDBInvoiceRepository{levelDB})

	return levelDB
}
