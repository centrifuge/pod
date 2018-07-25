// +build unit

package purchaseorderrepository

import (
	"testing"
	"os"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
)

var dbFileName = "/tmp/centrifuge_testing_podoc.leveldb"

func TestMain(m *testing.M) {
	defer Bootstrap().Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestStorageService(t *testing.T) {
	identifier := []byte("1")
	invalidIdentifier := []byte("2")

	purchaseorder := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier:identifier}}
	repo := GetPurchaseOrderRepository()
	err := repo.CreateOrUpdate(&purchaseorder)
	assert.Nil(t, err, "CreateOrUpdate should not return error")

	orderDocument, err := repo.FindById(identifier)
	assert.Nil(t, err, "FindById should not return error")
	assert.Equal(t, purchaseorder.CoreDocument.DocumentIdentifier, orderDocument.CoreDocument.DocumentIdentifier, "PurchaseOrder DocumentIdentifier should be equal")

	orderDocument, err = repo.FindById(invalidIdentifier)
	assert.NotNil(t, err, "FindById should not return error")
	assert.Nil(t, orderDocument, "PurchaseOrder should be NIL")
}

func TestLevelDBInvoiceRepository_StoreNilDocument(t *testing.T) {
	repo := GetPurchaseOrderRepository()
	err := repo.CreateOrUpdate(nil)

	assert.Error(t, err, "should have thrown an error")
}

func TestLevelDBInvoiceRepository_StoreNilCoreDocument(t *testing.T) {
	repo := GetPurchaseOrderRepository()
	err := repo.CreateOrUpdate(&purchaseorderpb.PurchaseOrderDocument{})

	assert.Error(t, err, "should have thrown an error")
}

func Bootstrap() (*leveldb.DB) {
	levelDB := storage.NewLeveldbStorage(dbFileName)

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	NewLevelDBPurchaseOrderRepository(&LevelDBPurchaseOrderRepository{levelDB})

	return levelDB
}
