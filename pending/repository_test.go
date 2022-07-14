//go:build unit
// +build unit

package pending

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var ctx map[string]interface{}
var cfg config.Configuration
var did = testingidentity.GenerateRandomDID()

func TestMain(m *testing.M) {
	ctx = make(map[string]interface{})
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobs.Bootstrapper{},
		&configstore.Bootstrapper{},
		&anchors.Bootstrapper{},
	}
	ctx[identity.BootstrappedDIDService] = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDFactory] = &identity.MockFactory{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", did.String())
	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func getRepository(ctx map[string]interface{}) Repository {
	db := ctx[storage.BootstrappedDB].(storage.Repository)
	return NewRepository(db)
}

type doc struct {
	documents.Document
	DocID, Current, Next []byte
	SomeString           string `json:"some_string"`
	Time                 time.Time
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

	repor.(*repo).db.Register(&doc{})
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

	assert.NoError(t, repor.Delete(accountID, id))
	m, err = repor.Get(accountID, id)
	assert.Error(t, err)
	assert.Nil(t, m)

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
