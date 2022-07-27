//go:build unit

package leveldb

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

type doc struct {
	Id         []byte `json:"id"`
	SomeString string `json:"some_string"`
}

func (m *doc) ID() ([]byte, error) {
	return m.Id, nil
}

func (m *doc) JSON() ([]byte, error) {
	return json.Marshal(m)
}

func (m *doc) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *doc) Type() reflect.Type {
	return reflect.TypeOf(m)
}

func getRandomRepository() (storage.Repository, string, error) {
	randomPath := GetRandomTestStoragePath()
	db, err := NewLevelDBStorage(randomPath)
	if err != nil {
		return nil, "", err
	}
	return NewLevelDBRepository(db), randomPath, nil
}

func TestNewLevelDBRepository(t *testing.T) {
	path := GetRandomTestStoragePath()
	db, err := NewLevelDBStorage(path)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	repo := NewLevelDBRepository(db)
	assert.NotNil(t, repo)
}

func TestLevelDBRepo_Register(t *testing.T) {
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	assert.Len(t, repo.(*levelDBRepo).models, 0, "should be empty")
	d := &doc{SomeString: "Hello, Repo!"}
	repo.Register(d)
	assert.Len(t, repo.(*levelDBRepo).models, 1, "should be not empty")
	assert.Contains(t, repo.(*levelDBRepo).models, "leveldb.doc")
}

func TestLevelDBRepo_Exists(t *testing.T) {
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	id := utils.RandomSlice(32)

	// Key doesnt exist
	assert.False(t, repo.Exists(id))

	d := &doc{SomeString: "Hello, Repo!"}
	err = repo.Create(id, d)
	assert.Nil(t, err)

	// Key exists
	assert.True(t, repo.Exists(id))
}

func TestLevelDBRepo_Get(t *testing.T) {
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	id := utils.RandomSlice(32)

	// Key doesnt exist
	_, err = repo.Get(id)
	assert.True(t, errors.IsOfType(storage.ErrModelRepositoryNotFound, err))

	d := &doc{SomeString: "Hello, Repo!"}
	err = repo.Create(id, d)
	assert.Nil(t, err)

	// Document not registered
	_, err = repo.Get(id)
	assert.True(t, errors.IsOfType(storage.ErrModelTypeNotRegistered, err))

	// Success
	repo.Register(&doc{})
	m, err := repo.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, d.SomeString, m.(*doc).SomeString)
}

func TestLevelDBRepo_GetAllByPrefix(t *testing.T) {
	prefix := "prefix-"
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	repo.Register(&doc{})

	// No match
	models, err := repo.GetAllByPrefix(prefix)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(models))

	id1 := append([]byte(prefix), utils.RandomSlice(32)...)
	id2 := append([]byte(prefix), utils.RandomSlice(32)...)
	d1 := &doc{SomeString: "Hello, Repo1!"}
	d2 := &doc{SomeString: "Hello, Repo2!"}
	err = repo.Create(id1, d1)
	assert.Nil(t, err)
	err = repo.Create(id2, d2)
	assert.Nil(t, err)

	models, err = repo.GetAllByPrefix(prefix)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(models))
}

func TestLevelDBRepo_Create(t *testing.T) {
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	id := utils.RandomSlice(32)

	d := &doc{SomeString: "Hello, Repo!"}
	err = repo.Create(id, d)
	assert.Nil(t, err)

	//Already exists
	err = repo.Create(id, d)
	assert.True(t, errors.IsOfType(storage.ErrRepositoryModelCreateKeyExists, err))
}

func TestLevelDBRepo_Update(t *testing.T) {
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	id := utils.RandomSlice(32)

	d := &doc{SomeString: "Hello, Repo!"}

	// Doesn't exist
	err = repo.Update(id, d)
	assert.True(t, errors.IsOfType(storage.ErrRepositoryModelUpdateKeyNotFound, err))

	err = repo.Create(id, d)
	assert.Nil(t, err)

	// Exists
	err = repo.Update(id, d)
	assert.Nil(t, err)
}

func TestLevelDBRepo_Delete(t *testing.T) {
	repo, _, err := getRandomRepository()
	assert.Nil(t, err)
	id := utils.RandomSlice(32)

	d := &doc{SomeString: "Hello, Repo!"}
	repo.Register(d)

	//Doesnt fail on key that doesnt exist
	err = repo.Delete(id)
	assert.Nil(t, err)

	err = repo.Create(id, d)
	assert.Nil(t, err)

	// Entry exists
	m, err := repo.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, d.SomeString, m.(*doc).SomeString)

	err = repo.Delete(id)
	assert.Nil(t, err)

	// Entry doesnt exist
	_, err = repo.Get(id)
	assert.True(t, errors.IsOfType(storage.ErrModelRepositoryNotFound, err))
}
