// +build unit

package storage

import (
	"testing"
	"bytes"
	"os"
)

var dbFileName = "/tmp/centrifuge_testing_storage.leveldb"
var storageDb LeveldbDataStore

func TestMain(m *testing.M) {
	storageDb = LeveldbDataStore{Path: dbFileName}
	storageDb.Open()
	defer storageDb.Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestLeveldbDataStore(t *testing.T) {
	instance = &storageDb

	one := []byte("1")
	two := []byte("2")
	three := []byte("3")

	instance.Put(one, two)
	instance.Put(two, two)
	instance.Put(two, three)

	get_one, err := instance.Get(one)
	if err != nil {
		t.Fatal("Exception when getting 'one", err)
	}

	if !bytes.Equal(two, get_one) {
		t.Fatal(two, "not equal", get_one)
	}

	get_two, err := instance.Get(two)
	if err != nil {
		t.Fatal("Exception when getting 'two'", err)
	}

	if !bytes.Equal(three, get_two) {
		t.Fatal(three, "not equal", get_two)
	}

	get_three, err := instance.Get(three)

	if err == nil {
		t.Fatal("Error should have been returned")
	}

	if get_three != nil {
		t.Fatal("Result from three should be nil")
	}

}