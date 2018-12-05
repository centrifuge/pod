// +build unit integration

package documents

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func getRepository(ctx map[string]interface{}) Repository {
	db := ctx[storage.BootstrappedLevelDB].(*leveldb.DB)
	return NewLevelDBRepository(db)
}

type doc struct {
	Model
	SomeString string `json:"some_string"`
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

func TestLevelDBRepo_Create_Exists(t *testing.T) {
	repo := getRepository(ctx)
	d := &doc{SomeString: "Hello, World!"}
	tenantID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	assert.False(t, repo.Exists(tenantID, id), "doc must not be present")
	err := repo.Create(tenantID, id, d)
	assert.Nil(t, err, "Create: unknown error")
	assert.True(t, repo.Exists(tenantID, id), "doc must be present")

	// overwrite
	err = repo.Create(tenantID, id, d)
	assert.Error(t, err, "Create: must not overwrite existing doc")
}

func TestLevelDBRepo_Update_Exists(t *testing.T) {
	repo := getRepository(ctx)
	d := &doc{SomeString: "Hello, World!"}
	tenantID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	assert.False(t, repo.Exists(tenantID, id), "doc must not be present")
	err := repo.Update(tenantID, id, d)
	assert.Error(t, err, "Update: should error out")
	assert.False(t, repo.Exists(tenantID, id), "doc must not be present")

	// overwrite
	err = repo.Create(tenantID, id, d)
	assert.Nil(t, err, "Create: unknown error")
	d.SomeString = "Hello, Repo!"
	err = repo.Update(tenantID, id, d)
	assert.Nil(t, err, "Update: unknown error")
	assert.True(t, repo.Exists(tenantID, id), "doc must be [resent")
}

func TestLevelDBRepo_Register(t *testing.T) {
	repo := getRepository(ctx)
	assert.Len(t, repo.(*levelDBRepo).models, 0, "should be empty")
	d := &doc{SomeString: "Hello, Repo!"}
	repo.Register(d)
	assert.Len(t, repo.(*levelDBRepo).models, 1, "should be not empty")
	assert.Contains(t, repo.(*levelDBRepo).models, "documents.doc")
}

func TestLevelDBRepo_Get_Create_Update(t *testing.T) {
	repo := getRepository(ctx)

	tenantID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	m, err := repo.Get(tenantID, id)
	assert.Error(t, err, "must return error")
	assert.Nil(t, m)

	d := &doc{SomeString: "Hello, Repo!"}
	err = repo.Create(tenantID, id, d)
	assert.Nil(t, err, "Create: unknown error")

	m, err = repo.Get(tenantID, id)
	assert.Error(t, err, "doc is not registered yet")
	assert.Nil(t, m)

	repo.Register(&doc{})
	m, err = repo.Get(tenantID, id)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	nd := m.(*doc)
	assert.Equal(t, d, nd, "must be equal")

	d.SomeString = "Hello, World!"
	err = repo.Update(tenantID, id, d)
	assert.Nil(t, err, "Update: unknown error")

	m, err = repo.Get(tenantID, id)
	assert.Nil(t, err, "Get: unknown error")
	nd = m.(*doc)
	assert.Equal(t, d, nd, "must be equal")
}
