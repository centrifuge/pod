// +build unit

package api

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
		"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
		"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	flag.Parse()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestCentAPIServer_StartHappy(t *testing.T) {
	//capi := NewCentAPIServer("0.0.0.0:9000", 9000, "")
	//ctx, canc := context.WithCancel(context.Background())
	//startErr := make(chan error)
	//go capi.Start(ctx, startErr)
	//err := <-startErr
	//fmt.Println(err)
	//canc()
	// TODO make this a proper test with an API health check call
}

func TestCentAPIServer_StartContextCancel(t *testing.T) {
	documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, invoice.DefaultService(nil, nil))
	capi := NewCentAPIServer("0.0.0.0:9000", 9000, "")
	ctx, canc := context.WithCancel(context.Background())
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	// TODO make some api call(healthcheck) to make sure that API is up
	// cancel the context to shutdown the server
	canc()
	wg.Wait()
	// TODO make some api call(healthcheck) to make sure that API is down, for now the fact that this test stops is enough to see that this is a success
}

func TestCentAPIServer_StartListenError(t *testing.T) {
	invoice.InitLegacyRepository(nil)
	// cause an error by using an invalid port
	capi := NewCentAPIServer("0.0.0.0:100000000", 100000000, "")
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
