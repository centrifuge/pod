// +build unit

package api

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		p2p.Bootstrapper{},
		&invoice.Bootstrapper{},
		&purchaseorder.Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestCentAPIServer_StartContextCancel(t *testing.T) {
	documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, invoice.DefaultService(nil, nil, nil))
	capi := apiServer{
		addr:    "0.0.0.0:9000",
		port:    9000,
		network: "",
	}
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
	capi := apiServer{
		addr:    "0.0.0.0:100000000",
		port:    100000000,
		network: "",
	}
	ctx, _ := context.WithCancel(context.Background())
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "listen tcp: address 100000000: invalid port", err.Error())
}
