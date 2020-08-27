// +build unit

package api

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/httpapi/userapi"
	"github.com/centrifuge/go-centrifuge/httpapi/v2"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient

	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient

	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
		&queue.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		pending.Bootstrapper{},
		&entityrelationship.Bootstrapper{},
		generic.Bootstrapper{},
		&ethereum.Bootstrapper{},
		&nft.Bootstrapper{},
		&queue.Starter{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		coreapi.Bootstrapper{},
		&entity.Bootstrapper{},
		funding.Bootstrapper{},
		transferdetails.Bootstrapper{},
		userapi.Bootstrapper{},
		v2.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)

	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestCentAPIServer_StartContextCancel(t *testing.T) {
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 9000)
	cfg.Set("centrifugeNetwork", "")
	capi := apiServer{config: cfg}
	ctx, canc := context.WithCancel(context.WithValue(context.Background(), bootstrap.NodeObjRegistry, ctx))
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	// cancel the context to shutdown the server
	canc()
	wg.Wait()
}

func TestCentAPIServer_StartListenError(t *testing.T) {
	// cause an error by using an invalid port
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 100000000)
	cfg.Set("centrifugeNetwork", "")
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), bootstrap.NodeObjRegistry, ctx))
	defer cancel()
	capi := apiServer{config: cfg}
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "listen tcp: address 100000000: invalid port", err.Error())
}

func TestCentAPIServer_FailedToGetRegistry(t *testing.T) {
	// cause an error by using an invalid port
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 100000000)
	cfg.Set("centrifugeNetwork", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	capi := apiServer{config: cfg}
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "failed to get NodeObjRegistry", err.Error())
}
