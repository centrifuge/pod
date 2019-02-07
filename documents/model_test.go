// +build unit

package documents

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/anchors"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/queue"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/assert"
)

var ctx map[string]interface{}
var ConfigService config.Service
var cfg config.Configuration

func TestMain(m *testing.M) {
	ctx = make(map[string]interface{})
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		transactions.Bootstrapper{},
		&queue.Bootstrapper{},
		&anchors.Bootstrapper{},
		&Bootstrapper{},
	}
	ctx[identity.BootstrappedIDService] = &testingcommons.MockIDService{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	ConfigService = ctx[config.BootstrappedConfigStorage].(config.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func Test_fetchUniqueCollaborators(t *testing.T) {

	tests := []struct {
		old    [][]byte
		new    []string
		result []identity.CentID
		err    bool
	}{
		{
			new:    []string{"0x010203040506"},
			result: []identity.CentID{{1, 2, 3, 4, 5, 6}},
		},

		{
			old:    [][]byte{{1, 2, 3, 2, 3, 1}},
			new:    []string{"0x010203040506"},
			result: []identity.CentID{{1, 2, 3, 4, 5, 6}},
		},

		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
			new: []string{"0x010203040506"},
		},

		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
		},

		// new collaborator with wrong format
		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
			new: []string{"0x0102030405"},
			err: true,
		},
	}

	for _, c := range tests {
		uc, err := fetchUniqueCollaborators(c.old, c.new)
		if err != nil {
			if c.err {
				continue
			}

			t.Fatal(err)
		}

		assert.Equal(t, c.result, uc)
	}
}

func TestCoreDocumentModel_PrepareNewVersion(t *testing.T) {
	dm := NewCoreDocModel()
	cd := dm.Document
	assert.NotNil(t, cd)

	//collaborators need to be hex string
	collabs := []string{"some ID"}
	newDocModel, err := dm.PrepareNewVersion(collabs)
	assert.Error(t, err)
	assert.Nil(t, newDocModel)

	// missing DocumentRoot
	c1 := utils.RandomSlice(6)
	c2 := utils.RandomSlice(6)
	c := []string{hexutil.Encode(c1), hexutil.Encode(c2)}
	ndm, err := dm.PrepareNewVersion(c)
	assert.NotNil(t, err)
	assert.Nil(t, ndm)

	// successful preparation of new version upon addition of DocumentRoot
	cd.DocumentRoot = utils.RandomSlice(32)
	ndm, err = dm.PrepareNewVersion(c)
	assert.Nil(t, err)
	assert.NotNil(t, ndm)

	// successful updating of version in new Document
	ncd := ndm.Document
	ocd := dm.Document
	assert.Equal(t, ncd.PreviousVersion, ocd.CurrentVersion)
	assert.Equal(t, ncd.CurrentVersion, ocd.NextVersion)

	// DocumentIdentifier has not changed
	assert.Equal(t, ncd.DocumentIdentifier, ocd.DocumentIdentifier)

	// DocumentRoot was updated
	assert.Equal(t, ncd.PreviousRoot, ocd.DocumentRoot)
}

func TestReadACLs_initReadRules(t *testing.T) {
	dm := NewCoreDocModel()
	cd := dm.Document
	err := dm.initReadRules(nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrZeroCollaborators, err))

	cs := []identity.CentID{identity.RandomCentID()}
	err = dm.initReadRules(cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)

	err = dm.initReadRules(cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)
}

func TestReadAccessValidator_AccountCanRead(t *testing.T) {
	dm := NewCoreDocModel()
	account, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	ndm, err := dm.PrepareNewVersion([]string{account.String()})
	cd := ndm.Document
	assert.NoError(t, err)
	assert.NotNil(t, cd.ReadRules)
	assert.NotNil(t, cd.Roles)

	// account who cant access
	rcid := identity.RandomCentID()
	assert.False(t, ndm.AccountCanRead(rcid))

	// account can access
	assert.True(t, ndm.AccountCanRead(account))
}

