// +build unit

package documents

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
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
		uc, err := FetchUniqueCollaborators(c.old, c.new)
		if err != nil {
			if c.err {
				continue
			}

			t.Fatal(err)
		}

		assert.Equal(t, c.result, uc)
	}
}
