// +build integration

package ethereum_test

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity/ethid"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration
var ctx = map[string]interface{}{}

var queueStartBootstrap bootstrap.TestBootstrapper

func TestMain(m *testing.M) {
	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		transactions.Bootstrapper{},
		ethereum.Bootstrapper{},
		&queue.Bootstrapper{},
		&ethid.Bootstrapper{},
		&configstore.Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)

	queueStartBootstrap = &queue.Starter{}
	bootstappers = append(bootstappers, queueStartBootstrap)

	result := m.Run()
	bootstrap.RunTestTeardown(bootstappers)
	os.Exit(result)
}

func bootstrapQueueStart() {
	queueStartBootstrap.TestBootstrap(ctx)

}

func TestGetConnection_returnsSameConnection(t *testing.T) {
	howMany := 5
	confChannel := make(chan ethereum.Client, howMany)
	for ix := 0; ix < howMany; ix++ {
		go func() {
			confChannel <- ethereum.GetClient()
		}()
	}
	for ix := 0; ix < howMany; ix++ {
		multiThreadCreatedCon := <-confChannel
		assert.Equal(t, multiThreadCreatedCon, ethereum.GetClient(), "Should only return a single ethereum client")
	}
}

func TestNewGethClient(t *testing.T) {
	gc, err := ethereum.NewGethClient(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, gc)
}
