//go:build integration

package documents

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
}

var (
	storageRepo storage.Repository
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	storageRepo = ctx[storage.BootstrappedDB].(storage.Repository)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

type doc struct {
	Document
	DocID, Current, Next []byte
	SomeString           string `json:"some_string"`
	Time                 time.Time
	status               Status
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
	return m.Current
}

func (m *doc) NextVersion() []byte {
	return m.Next
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
	return m.Time, nil
}

func (m *doc) GetStatus() Status {
	return m.status
}

func TestRepo_Create_Exists(t *testing.T) {
	repo := NewDBRepository(storageRepo)

	accountID, id := utils.RandomSlice(32), utils.RandomSlice(32)

	d := &doc{SomeString: "Hello, World!", DocID: id, status: Committed}

	assert.False(t, repo.Exists(accountID, id), "doc must not be present")

	err := repo.Create(accountID, id, d)
	assert.Nil(t, err, "Create: unknown error")
	assert.True(t, repo.Exists(accountID, id), "doc must be present")

	// overwrite
	err = repo.Create(accountID, id, d)
	assert.Error(t, err, "Create: must not overwrite existing doc")
}

func TestRepo_Update_Exists(t *testing.T) {
	repo := NewDBRepository(storageRepo)

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
	assert.True(t, repo.Exists(accountID, id), "doc must be present")
}

func TestRepo_Get_Create_Update(t *testing.T) {
	repo := NewDBRepository(storageRepo)

	accountID, id := utils.RandomSlice(32), utils.RandomSlice(32)
	m, err := repo.Get(accountID, id)
	assert.Error(t, err, "must return error")
	assert.Nil(t, m)

	d := &doc{SomeString: "Hello, Repo!", DocID: id}
	err = repo.Create(accountID, id, d)
	assert.Nil(t, err, "Create: unknown error")

	m, err = repo.Get(accountID, id)
	assert.Error(t, err, "doc is not registered yet")
	assert.Nil(t, m)

	repo.Register(&doc{})

	m, err = repo.Get(accountID, id)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	nd := m.(*doc)
	assert.Equal(t, d, nd, "must be equal")

	d.SomeString = "Hello, World!"

	err = repo.Update(accountID, id, d)
	assert.Nil(t, err, "Update: unknown error")

	m, err = repo.Get(accountID, id)
	assert.Nil(t, err, "Get: unknown error")

	nd = m.(*doc)
	assert.Equal(t, d, nd, "must be equal")

	// a document id sent which is not a model
	storageRepo.Register(&unknownDoc{})
	unid := utils.RandomSlice(32)
	u := unknownDoc{SomeString: "unknown"}

	err = storageRepo.Create(getKey(accountID, unid), &u)
	assert.NoError(t, err)

	_, err = repo.Get(accountID, unid)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is not a model object")
	}
}

func TestIntegration_Repo_GetLatest(t *testing.T) {
	// missing latest key
	acc := utils.RandomSlice(20)
	id := utils.RandomSlice(32)

	r := NewDBRepository(storageRepo)

	_, err := r.GetLatest(acc, id)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(storage.ErrModelRepositoryNotFound, err))

	// different type
	rr := r.(*repo)

	r.Register(new(doc))

	err = rr.db.Create(getLatestKey(acc, id), new(doc))
	assert.NoError(t, err)

	_, err = r.GetLatest(acc, id)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))

	// missing version
	rr.db.Register(new(latestVersion))

	lv := &latestVersion{
		CurrentVersion: utils.RandomSlice(32),
	}
	err = rr.db.Create(getLatestKey(acc, id), lv)
	assert.NoError(t, err)

	_, err = r.GetLatest(acc, id)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(storage.ErrModelRepositoryNotFound, err))
	assert.True(t, rr.db.Exists(getLatestKey(acc, id)))

	// success
	d := new(doc)

	err = rr.db.Create(getKey(acc, lv.CurrentVersion), d)
	assert.NoError(t, err)

	m, err := r.GetLatest(acc, id)
	assert.NoError(t, err)
	assert.Equal(t, d, m)
}

func TestRepo_updateLatestIndex(t *testing.T) {
	r := NewDBRepository(storageRepo)

	rr := r.(*repo)

	acc := utils.RandomSlice(20)
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	tm := time.Now().UTC()

	// missing index, should create one
	d := &doc{
		DocID:   id,
		Current: id,
		Next:    next,
		Time:    tm,
		status:  Committed,
	}
	assert.False(t, rr.db.Exists(getLatestKey(acc, id)))
	err := rr.updateLatestIndex(acc, d)

	assert.NoError(t, err)
	assert.True(t, rr.db.Exists(getLatestKey(acc, id)))

	lv, err := rr.getLatestVersion(getLatestKey(acc, id))
	assert.NoError(t, err)
	assert.Equal(t, &latestVersion{
		CurrentVersion: id,
		Timestamp:      tm,
		NextVersion:    next,
	}, lv)

	// next version
	d.Current = next
	d.Next = utils.RandomSlice(32)
	d.Time = time.Now().UTC()

	err = rr.updateLatestIndex(acc, d)
	assert.NoError(t, err)
	assert.True(t, rr.db.Exists(getLatestKey(acc, id)))

	lv, err = rr.getLatestVersion(getLatestKey(acc, id))
	assert.NoError(t, err)
	assert.Equal(t, &latestVersion{
		CurrentVersion: next,
		Timestamp:      d.Time,
		NextVersion:    d.Next,
	}, lv)

	// later time
	d.Current = utils.RandomSlice(32)
	d.Next = utils.RandomSlice(32)
	tm = time.Now().UTC()
	assert.False(t, d.Time.Equal(tm))
	d.Time = tm

	err = rr.updateLatestIndex(acc, d)
	assert.NoError(t, err)
	assert.True(t, rr.db.Exists(getLatestKey(acc, id)))

	lv, err = rr.getLatestVersion(getLatestKey(acc, id))
	assert.NoError(t, err)
	assert.Equal(t, &latestVersion{
		CurrentVersion: d.Current,
		Timestamp:      d.Time,
		NextVersion:    d.Next,
	}, lv)

	// older version, dont update index
	d.Time = time.Now().UTC().Add(-time.Hour)
	oldC := d.Current
	oldN := d.Next
	d.Current = utils.RandomSlice(32)
	d.Next = utils.RandomSlice(32)

	err = rr.updateLatestIndex(acc, d)
	assert.NoError(t, err)
	assert.True(t, rr.db.Exists(getLatestKey(acc, id)))

	lv, err = rr.getLatestVersion(getLatestKey(acc, id))
	assert.NoError(t, err)
	assert.Equal(t, &latestVersion{
		CurrentVersion: oldC,
		Timestamp:      tm,
		NextVersion:    oldN,
	}, lv)
}
