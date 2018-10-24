// +build unit

package invoice

import (
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&anchors.Bootstrapper{},
		&Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(ibootstappers, nil)
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
