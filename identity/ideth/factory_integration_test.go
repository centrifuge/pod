// +build integration

package ideth

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
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
		jobsv2.Bootstrapper{},
		&queue.Bootstrapper{},
		ethereum.Bootstrapper{},
		&Bootstrapper{},
		&configstore.Bootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	dispatcher := ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	ctxh, canc := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go dispatcher.Start(ctxh, wg, nil)
	result := m.Run()
	canc()
	bootstrap.RunTestTeardown(bootstappers)
	wg.Wait()
	os.Exit(result)
}

func TestCreateIdentity_successful(t *testing.T) {
	DeployIdentity(t, ctx, cfg)
}
