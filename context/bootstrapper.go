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
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/version"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("context")

// MainBootstrapper holds all the bootstrapper implementations
type MainBootstrapper struct {
	Bootstrappers []bootstrap.Bootstrapper
}

// PopulateBaseBootstrappers adds all the bootstrapper implementations to MainBootstrapper
func (m *MainBootstrapper) PopulateBaseBootstrappers() {
	m.Bootstrappers = []bootstrap.Bootstrapper{
		&version.Bootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		ethereum.Bootstrapper{},
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

// PopulateRunBootstrappers adds blocking Node bootstrapper at the end.
// Note: Node bootstrapper must be the last bootstrapper to be invoked as it won't return until node is shutdown
func (m *MainBootstrapper) PopulateRunBootstrappers() {
	m.PopulateBaseBootstrappers()
	m.Bootstrappers = append(m.Bootstrappers, &node.Bootstrapper{})
}

// Bootstrap runs all the loaded bootstrapper implementations.
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

// TenantBootstrapper ensures all dependencies for a tenant is initialised correctly
type TenantBootstrapper struct {
	IOCContext    map[string]interface{}
	Bootstrappers []bootstrap.Bootstrapper
}

func (t *TenantBootstrapper) Populate() {
	t.Bootstrappers = []bootstrap.Bootstrapper{
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		&invoice.Bootstrapper{},
		&purchaseorder.Bootstrapper{},
		&nft.Bootstrapper{},
	}
}

// Bootstrap runs all the loaded bootstrapper implementations.
// The context passed has to be the same one used for the MainBootstrapper otherwise this will error out
func (t *TenantBootstrapper) Bootstrap(context map[string]interface{}) error {
	t.IOCContext = make(map[string]interface{})
	// make a copy of the context
	for k, v := range context {
		t.IOCContext[k] = v
	}
	for _, b := range t.Bootstrappers {
		err := b.Bootstrap(t.IOCContext)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			return err
		}
	}
	return nil
}
