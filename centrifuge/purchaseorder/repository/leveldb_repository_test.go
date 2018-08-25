// +build integration

package purchaseorderrepository

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

var dbFileName = "/tmp/centrifuge_testing_podoc.leveldb"

func TestMain(m *testing.M) {
	levelDB := storage.NewLevelDBStorage(dbFileName)
	coredocumentrepository.NewLevelDBRepository(&coredocumentrepository.LevelDBRepository{LevelDB: levelDB})
	InitLevelDBRepository(levelDB)
	result := m.Run()
	levelDB.Close()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestStorageService(t *testing.T) {
	identifier := []byte("1")
	invalidIdentifier := []byte("2")

	purchaseorder := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}}
	repo := GetRepository()
	err := repo.Create(identifier, &purchaseorder)
	assert.Nil(t, err, "Create should not return error")

	orderDocument := &purchaseorderpb.PurchaseOrderDocument{}
	err = repo.GetByID(identifier, orderDocument)
	assert.Nil(t, err, "GetByID should not return error")
	assert.Equal(t, purchaseorder.CoreDocument.DocumentIdentifier, orderDocument.CoreDocument.DocumentIdentifier, "PurchaseOrder DocumentIdentifier should be equal")

	err = repo.GetByID(invalidIdentifier, orderDocument)
	assert.NotNil(t, err, "FindById should not return error")
}

func TestLevelDBPurchaseRepository_StoreNilDocument(t *testing.T) {
	repo := GetRepository()
	err := repo.Create([]byte("1"), nil)
	assert.Error(t, err, "should have thrown an error")
}

func TestLevelDBPurchaseRepository_StoreNilCoreDocument(t *testing.T) {
	repo := GetRepository()
	err := repo.Create([]byte("1"), &purchaseorderpb.PurchaseOrderDocument{})
	assert.Error(t, err, "should have thrown an error")
}
