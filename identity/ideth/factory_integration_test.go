// +build integration

package ideth

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration
var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {

	ctx = testingutils.BuildIntegrationTestingContext()

	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
		&queue.Bootstrapper{},
		ethereum.Bootstrapper{},
		&Bootstrapper{},
		&configstore.Bootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(bootstappers)
	os.Exit(result)
}

func TestCreateIdentity_successful(t *testing.T) {
	factory := ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	accountCtx := testingconfig.CreateAccountContext(t, cfg)
	did, err := factory.CreateIdentity(accountCtx)
	assert.Nil(t, err, "create identity should be successful")

	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.ToAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")
	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be around 3378 bytes")
}
