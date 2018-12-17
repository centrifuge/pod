// +build integration

package testingbootstrap

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("context")

var bootstappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&storage.Bootstrapper{},
	ethereum.Bootstrapper{},
	&queue.Bootstrapper{},
	anchors.Bootstrapper{},
	&identity.Bootstrapper{},
	documents.Bootstrapper{},
	p2p.Bootstrapper{},
	&invoice.Bootstrapper{},
	&purchaseorder.Bootstrapper{},
	&nft.Bootstrapper{},
	&queue.Starter{},
}

func TestFunctionalEthereumBootstrap() map[string]interface{} {
	ctx := map[string]interface{}{}
	for _, b := range bootstappers {
		err := b.TestBootstrap(ctx)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}

	return ctx
}
func TestFunctionalEthereumTearDown() {
	for _, b := range bootstappers {
		err := b.TestTearDown()
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}
