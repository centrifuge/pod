package context

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/api"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/p2p"
		"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/version"
	logging "github.com/ipfs/go-log"
	"github.com/centrifuge/go-centrifuge/queue"
)

var log = logging.Logger("context")

type MainBootstrapper struct {
	Bootstrappers []bootstrap.Bootstrapper
}

func (m *MainBootstrapper) PopulateBaseBootstrappers() {
	m.Bootstrappers = []bootstrap.Bootstrapper{
		&version.Bootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&ethereum.Bootstrapper{},
		&queue.Bootstrapper{},
		&anchors.Bootstrapper{},
		&identity.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		api.Bootstrapper{},
		&invoice.Bootstrapper{},
		&purchaseorder.Bootstrapper{},
		&nft.Bootstrapper{},
	}
}

func (m *MainBootstrapper) PopulateRunBootstrappers() {
	m.PopulateBaseBootstrappers()
	// NODE BOOTSTRAPPER MUST BE THE LAST BOOTSTRAPPER TO BE INVOKED AS IT WON'T RETURN UNTIL NODE IS SHUTDOWN
	m.Bootstrappers = append(m.Bootstrappers, &node.Bootstrapper{})
}

func (m *MainBootstrapper) Bootstrap(context map[string]interface{}) error {
	for _, b := range m.Bootstrappers {
		err := b.Bootstrap(context)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			return err
		}
	}
	return nil
}
