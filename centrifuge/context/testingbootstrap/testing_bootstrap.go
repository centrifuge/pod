package testingbootstrap

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
		"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	logging "github.com/ipfs/go-log"
	"github.com/centrifuge/go-centrifuge/centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/purchaseorder"
)

var log = logging.Logger("context")

// ---- Ethereum ----
func TestFunctionalEthereumBootstrap() {
	bootstrapQueuing()
}
func TestFunctionalEthereumTearDown() {
	tearDownQueuing()
}

// ---- END Ethereum ----

// ---- Integration Testing ----

var ibootstappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&storage.Bootstrapper{},
	&coredocumentrepository.Bootstrapper{},
	&invoice.Bootstrapper{},
	&purchaseorder.Bootstrapper{},
}

func TestIntegrationBootstrap() {
	contextval := map[string]interface{}{}
	for _, b := range ibootstappers {
		err := b.TestBootstrap(contextval)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}

func TestIntegrationTearDown() {
	for _, b := range ibootstappers {
		err := b.TestTearDown()
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}

var bootstappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&storage.Bootstrapper{},
	&coredocumentrepository.Bootstrapper{},
	&invoice.Bootstrapper{},
	&purchaseorder.Bootstrapper{},
	&ethereum.Bootstrapper{},
	&anchors.Bootstrapper{},
	&identity.Bootstrapper{},
	&queue.Bootstrapper{},
}

// ---- End Integration Testing ----
func bootstrapQueuing() {
	// TODO here we would not have to put the bootstrapper.BootstrappedConfig after the TestBootstrapper refactoring
	contextval := map[string]interface{}{}
	for _, b := range bootstappers {
		err := b.TestBootstrap(contextval)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}

func tearDownQueuing() {
	for _, b := range bootstappers {
		err := b.TestTearDown()
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}
