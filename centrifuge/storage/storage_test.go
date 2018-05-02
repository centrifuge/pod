// +build unit

package storage

import (
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

var dbFileName = "/tmp/centrifuge_testing_storage.leveldb"
var storageDb *leveldb.DB

func TestMain(m *testing.M) {
	storageDb = NewLeveldbStorage(dbFileName)
	defer storageDb.Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestGetLeveldbStorage(t *testing.T) {
	one := []byte("1")
	two := []byte("2")

	err := storageDb.Put(one, two, nil)
	assert.Nil(t, err, "Should not error out")

	get_one, err := storageDb.Get(one, nil)
	assert.Nil(t, err, "Should not error out")
	assert.Equal(t, two, get_one)
}