// +build integration

package cmd

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/did"
	"github.com/centrifuge/go-centrifuge/identity/ethid"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration
var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		transactions.Bootstrapper{},
		&queue.Bootstrapper{},
		ethereum.Bootstrapper{},
		&ethid.Bootstrapper{},
		&configstore.Bootstrapper{},
		&did.Bootstrapper{},
		&queue.Starter{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(bootstappers)
	os.Exit(result)
}

func TestCreateConfig(t *testing.T) {
	dataDir := "testconfig"
	keyPath := path.Join(testingutils.GetProjectDir(), "build/scripts/test-dependencies/test-ethereum/migrateAccount.json")
	scAddrs := did.GetSmartContractAddresses()
	err := CreateConfig(dataDir, "http://127.0.0.1:9545", keyPath, "", "russianhill", 8028, 38202, nil, true, "", scAddrs)
	assert.Nil(t, err, "Create Config should be successful")

	//TODO check file and identity contract

	err = exec.Command("rm", "-rf", dataDir).Run()
	assert.Nil(t, err, "removing testconfig folder should be successful")

}
