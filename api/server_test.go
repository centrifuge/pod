// +build unit

package api

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg *config.Configuration
var registry *documents.ServiceRegistry

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		ethereum.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		&invoice.Bootstrapper{},
		&purchaseorder.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(*config.Configuration)
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestCentAPIServer_StartContextCancel(t *testing.T) {
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 9000)
	cfg.Set("centrifugeNetwork", "")
	registry.Register(documenttypes.InvoiceDataTypeUrl, invoice.DefaultService(cfg, nil, nil, nil, nil))
	capi := apiServer{config: cfg, registry: registry}
	ctx, canc := context.WithCancel(context.Background())
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
	ctx, _ := context.WithCancel(context.Background())
	capi := apiServer{config: cfg, registry: registry}
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "listen tcp: address 100000000: invalid port", err.Error())
}
