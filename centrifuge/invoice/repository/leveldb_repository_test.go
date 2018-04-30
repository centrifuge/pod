// +build unit

package repository

import (
	"testing"
	"os"
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/stretchr/testify/assert"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
)

var dbFileName = "/tmp/centrifuge_testing_invdoc.leveldb"

func TestMain(m *testing.M) {
	cc.Bootstrap()
	defer cc.LevelDB.Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestStorageService(t *testing.T) {
	identifier := []byte("1")
	invalidIdentifier := []byte("2")

	invoice := invoicepb.InvoiceDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier:identifier}}
	repo := NewLevelDBInvoiceRepository(cc.LevelDB)
	err := repo.Store(&invoice)
	assert.Nil(t, err, "Store should not return error")

	inv, err := repo.FindById(identifier)
	assert.Nil(t, err, "FindById should not return error")
	assert.Equal(t, invoice.CoreDocument.DocumentIdentifier, inv.CoreDocument.DocumentIdentifier, "Invoice DocumentIdentifier should be equal")

	inv, err = repo.FindById(invalidIdentifier)
	assert.NotNil(t, err, "FindById should not return error")
	assert.Nil(t, inv, "Invoice should be NIL")
}
