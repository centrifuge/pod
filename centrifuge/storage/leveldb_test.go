// +build unit

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
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
	id := tools.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id}}
	err := defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")

	err = defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Error(t, err, "create must fail")

	defaultDB.ValidateFunc = func(proto.Message) error {
		return fmt.Errorf("failed validation")
	}

	id2 := tools.RandomSlice(32)
	order.CoreDocument.DocumentIdentifier = id2
	err = defaultDB.Create(id2, &order)
	assert.Error(t, err, "create must fail")
	defaultDB.ValidateFunc = nil
}

func TestDefaultLevelDB_Exists(t *testing.T) {
	id1 := tools.RandomSlice(32)
	id2 := tools.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id1}}
	err := defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")

	assert.True(t, defaultDB.Exists(id1))
	assert.False(t, defaultDB.Exists(id2))
}

func TestDefaultLevelDB_GetKey(t *testing.T) {
	id := tools.RandomSlice(32)
	assert.Equal(t, id, defaultDB.GetKey(id))
}

func TestDefaultLevelDB_GetByID(t *testing.T) {
	id := tools.RandomSlice(32)
	order := purchaseorderpb.PurchaseOrderDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: id}}
	err := defaultDB.Create(order.CoreDocument.DocumentIdentifier, &order)
	assert.Nil(t, err, "create must pass")
	order2 := new(purchaseorderpb.PurchaseOrderDocument)
	assert.Nil(t, defaultDB.GetByID(id, order2), "error should be nil")
	assert.Equal(t, id, order2.CoreDocument.DocumentIdentifier)
}

func TestDefaultLevelDB_Update(t *testing.T) {
	id := tools.RandomSlice(32)
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

type model struct {
	shouldError bool
	Data        string `json:"data"`
}

func (m model) MarshalJSON() ([]byte, error) {
	if m.shouldError {
		return nil, fmt.Errorf("failed to marshal")
	}

	var d struct {
		Data string
	}
	d.Data = m.Data
	return json.Marshal(d)
}

func (m *model) UnmarshalJSON(data []byte) error {
	if m.shouldError {
		return fmt.Errorf("failed to unmarshal")
	}

	var d struct {
		Data string
	}

	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	m.Data = d.Data
	return nil
}

func TestDefaultLevelDB_GetModelByID_missing(t *testing.T) {
	id := tools.RandomSlice(32)
	m := new(model)
	err := defaultDB.GetModelByID(id, m)
	assert.Error(t, err, "error must be non nil")
}

func TestDefaultLevelDB_GetModelByID_nil_model(t *testing.T) {
	id := tools.RandomSlice(32)
	err := defaultDB.GetModelByID(id, nil)
	assert.Error(t, err, "error must be non nil")
}

func TestDefaultLevelDB_GetModelByID_Unmarshal_fail(t *testing.T) {
	id := tools.RandomSlice(32)
	m := &model{shouldError: true}
	err := defaultDB.GetModelByID(id, m)
	assert.Error(t, err, "error must be non nil")
}

func TestDefaultLevelDB_GetModelByID(t *testing.T) {
	id := tools.RandomSlice(32)
	m := &model{Data: "hello, world"}
	err := defaultDB.CreateModel(id, m)
	assert.Nil(t, err, "error should be nil")
	nm := new(model)
	err = defaultDB.GetModelByID(id, nm)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, m, nm, "models must match")
}

func TestDefaultLevelDB_CreateModel(t *testing.T) {
	id := tools.RandomSlice(32)
	d := &model{Data: "Create it"}
	defaultDB.ValidateModelFunc = func(m json.Marshaler) error { return nil }
	err := defaultDB.CreateModel(id, d)
	assert.Nil(t, err, "create must pass")

	// same id
	err = defaultDB.CreateModel(id, new(model))
	assert.Error(t, err, "create must fail")

	// nil model
	err = defaultDB.CreateModel(id, nil)
	assert.Error(t, err, "create must fail")

	// failed validation
	defaultDB.ValidateModelFunc = func(m json.Marshaler) error {
		return fmt.Errorf("failed validation")
	}
	err = defaultDB.CreateModel(tools.RandomSlice(32), new(model))
	assert.Error(t, err, "create must fail")
	defaultDB.ValidateModelFunc = nil
}

func TestDefaultLevelDB_UpdateModel(t *testing.T) {
	id := tools.RandomSlice(32)

	// missing Id
	err := defaultDB.UpdateModel(id, new(model))
	assert.Error(t, err, "update must fail")

	// nil model
	err = defaultDB.CreateModel(id, nil)
	assert.Error(t, err, "update must fail")

	m := &model{Data: "create it"}
	err = defaultDB.CreateModel(id, m)
	assert.Nil(t, err, "create must pass")

	// failed validation
	defaultDB.ValidateModelFunc = func(m json.Marshaler) error {
		return fmt.Errorf("failed validation")
	}
	err = defaultDB.CreateModel(tools.RandomSlice(32), new(model))
	assert.Error(t, err, "create must fail")
	defaultDB.ValidateModelFunc = nil

	// successful one
	m.Data = "update it"
	defaultDB.ValidateModelFunc = func(m json.Marshaler) error { return nil }
	err = defaultDB.UpdateModel(id, m)
	assert.Nil(t, err, "update must pass")
	nm := new(model)
	err = defaultDB.GetModelByID(id, nm)
	assert.Nil(t, err, "get mode must pass")
	assert.Equal(t, m, nm)
	defaultDB.ValidateModelFunc = nil
}
