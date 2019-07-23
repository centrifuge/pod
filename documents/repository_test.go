// +build unit

package documents

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func getRepository(ctx map[string]interface{}) Repository {
	db := ctx[storage.BootstrappedDB].(storage.Repository)
	return NewDBRepository(db)
}

type doc struct {
	Model
	DocID      []byte
	SomeString string `json:"some_string"`
}

type unknownDoc struct {
	SomeString string `json:"some_string"`
}

func (unknownDoc) Type() reflect.Type {
	return reflect.TypeOf(unknownDoc{})
}

func (u *unknownDoc) JSON() ([]byte, error) {
	return json.Marshal(u)
}

func (u *unknownDoc) FromJSON(j []byte) error {
	return json.Unmarshal(j, u)
}

func (m *doc) ID() []byte {
	return m.DocID
}

func (m *doc) CurrentVersion() []byte {
	return m.DocID
}

func (m *doc) NextVersion() []byte {
	return m.DocID
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

func (m *doc) Timestamp() (time.Time, error) {
	return time.Now().UTC(), nil
}

func TestLevelDBRepo_Create_Exists(t *testing.T) {
	repo := getRepository(ctx)
	accountID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	d := &doc{SomeString: "Hello, World!", DocID: id}
	assert.False(t, repo.Exists(accountID, id), "doc must not be present")
	err := repo.Create(accountID, id, d)
	assert.Nil(t, err, "Create: unknown error")
	assert.True(t, repo.Exists(accountID, id), "doc must be present")

	// overwrite
	err = repo.Create(accountID, id, d)
	assert.Error(t, err, "Create: must not overwrite existing doc")
}

func TestLevelDBRepo_Update_Exists(t *testing.T) {
	repo := getRepository(ctx)
	accountID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	d := &doc{SomeString: "Hello, World!", DocID: id}
	assert.False(t, repo.Exists(accountID, id), "doc must not be present")
	err := repo.Update(accountID, id, d)
	assert.Error(t, err, "Update: should error out")
	assert.False(t, repo.Exists(accountID, id), "doc must not be present")

	// overwrite
	err = repo.Create(accountID, id, d)
	assert.Nil(t, err, "Create: unknown error")
	d.SomeString = "Hello, Repo!"
	err = repo.Update(accountID, id, d)
	assert.Nil(t, err, "Update: unknown error")
	assert.True(t, repo.Exists(accountID, id), "doc must be [resent")
}

func TestLevelDBRepo_Get_Create_Update(t *testing.T) {
	repor := getRepository(ctx)

	accountID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	m, err := repor.Get(accountID, id)
	assert.Error(t, err, "must return error")
	assert.Nil(t, m)

	d := &doc{SomeString: "Hello, Repo!", DocID: id}
	err = repor.Create(accountID, id, d)
	assert.Nil(t, err, "Create: unknown error")

	m, err = repor.Get(accountID, id)
	assert.Error(t, err, "doc is not registered yet")
	assert.Nil(t, m)

	repor.Register(&doc{})
	m, err = repor.Get(accountID, id)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	nd := m.(*doc)
	assert.Equal(t, d, nd, "must be equal")

	d.SomeString = "Hello, World!"
	err = repor.Update(accountID, id, d)
	assert.Nil(t, err, "Update: unknown error")

	m, err = repor.Get(accountID, id)
	assert.Nil(t, err, "Get: unknown error")
	nd = m.(*doc)
	assert.Equal(t, d, nd, "must be equal")

	// a document id sent which is not a model
	repor.(*repo).db.Register(&unknownDoc{})
	unid := utils.RandomSlice(32)
	u := unknownDoc{SomeString: "unknown"}
	//hexKey := hexutil.Encode(append(accountID, unid...))
	err = repor.(*repo).db.Create(repor.(*repo).getKey(accountID, unid), &u)
	assert.NoError(t, err)
	m, err = repor.Get(accountID, unid)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is not a model object")
	}
}
