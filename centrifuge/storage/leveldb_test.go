// +build unit

package storage

import (
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

var dbFileName = "/tmp/centrifuge_testing_storage.leveldb"
var storageDb *leveldb.DB
var defaultDB *DefaultLevelDB

func TestMain(m *testing.M) {
	storageDb = NewLevelDBStorage(dbFileName)
	defer storageDb.Close()
	defaultDB = &DefaultLevelDB{LevelDB: storageDb}

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestGetLevelDBStorage(t *testing.T) {
	one := []byte("1")
	two := []byte("2")

	err := storageDb.Put(one, two, nil)
	assert.Nil(t, err, "Should not error out")

	getOnes, err := storageDb.Get(one, nil)
	assert.Nil(t, err, "Should not error out")
	assert.Equal(t, two, getOnes)
}

func TestDefaultLevelDB_Create(t *testing.T) {
	id := utils.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id}}
	err := defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")

	err = defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Error(t, err, "create must fail")

	defaultDB.ValidateFunc = func(proto.Message) error {
		return fmt.Errorf("failed validation")
	}

	id2 := utils.RandomSlice(32)
	order.CoreDocument.DocumentIdentifier = id2
	err = defaultDB.Create(id2, &order)
	assert.Error(t, err, "create must fail")
	defaultDB.ValidateFunc = nil
}

func TestDefaultLevelDB_Exists(t *testing.T) {
	id1 := utils.RandomSlice(32)
	id2 := utils.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id1}}
	err := defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")

	assert.True(t, defaultDB.Exists(id1))
	assert.False(t, defaultDB.Exists(id2))
}

func TestDefaultLevelDB_GetKey(t *testing.T) {
	id := utils.RandomSlice(32)
	assert.Equal(t, id, defaultDB.GetKey(id))
}

func TestDefaultLevelDB_GetByID(t *testing.T) {
	id := utils.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id}}
	err := defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")
	order2 := new(purchaseorderpb.PurchaseOrderDocument)
	assert.Nil(t, defaultDB.GetByID(id, order2), "error should be nil")
	assert.Equal(t, id, order2.CoreDocument.DocumentIdentifier)
}

func TestDefaultLevelDB_Update(t *testing.T) {
	id := utils.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id}}
	err := defaultDB.Update(order.CoreDocument.DocumentIdentifier, &order)
	assert.Error(t, err, "create must fail")

	err = defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")

	err = defaultDB.Update(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "update must pass")

	defaultDB.ValidateFunc = func(proto.Message) error {
		return fmt.Errorf("failed validation")
	}

	err = defaultDB.Update(order.CoreDocument.DocumentIdentifier, &order)
	assert.Error(t, err, "update must fail")

	defaultDB.ValidateFunc = nil
}
