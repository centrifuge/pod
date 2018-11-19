// +build unit

package invoice

import (
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
)

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx := map[string]interface{}{
		bootstrap.BootstrappedEthereumClient: ethClient,
	}
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		&Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestBootstrapper_Bootstrap(t *testing.T) {
	//err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	//assert.Error(t, err, "Should throw an error because of empty context")
}

func TestBootstrapper_registerInvoiceService(t *testing.T) {
	//context := map[string]interface{}{}
	//context[bootstrap.BootstrappedLevelDb] = true
	//err := (&Bootstrapper{}).Bootstrap(context)
	//assert.Nil(t, err, "Should throw because context is passed")
	//
	////coreDocument embeds a invoice
	//coreDocument := testingutils.GenerateCoreDocument()
	//registry := documents.GetRegistryInstance()
	//
	//documentType, err := cd.GetTypeUrl(coreDocument)
	//assert.Nil(t, err, "should not throw an error because document contains a type")
	//
	//service, err := registry.LocateService(documentType)
	//assert.Nil(t, err, "service should be successful registered and able to locate")
	//assert.NotNil(t, service, "service should be returned")
}
