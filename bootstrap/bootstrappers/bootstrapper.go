package bootstrappers

import (
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/dispatcher"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/documents/entity"
	"github.com/centrifuge/pod/documents/entityrelationship"
	"github.com/centrifuge/pod/documents/generic"
	"github.com/centrifuge/pod/http"
	httpv2 "github.com/centrifuge/pod/http/v2"
	httpv3 "github.com/centrifuge/pod/http/v3"
	identityv2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/ipfs"
	"github.com/centrifuge/pod/jobs"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/node"
	"github.com/centrifuge/pod/p2p"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pending"
	"github.com/centrifuge/pod/storage/leveldb"
	"github.com/centrifuge/pod/version"
	log2 "github.com/ipfs/go-log"
)

var log = log2.Logger("context")

// MainBootstrapper holds all the bootstrapper implementations
type MainBootstrapper struct {
	Bootstrappers []bootstrap.Bootstrapper
}

// PopulateBaseBootstrappers adds all the bootstrapper implementations to MainBootstrapper
func (m *MainBootstrapper) PopulateBaseBootstrappers() {
	m.Bootstrappers = []bootstrap.Bootstrapper{
		&version.Bootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&jobs.Bootstrapper{},
		centchain.Bootstrapper{},
		&pallets.Bootstrapper{},
		&dispatcher.Bootstrapper{},
		&identityv2.Bootstrapper{},
		documents.Bootstrapper{},
		&http.Bootstrapper{},
		&entityrelationship.Bootstrapper{},
		generic.Bootstrapper{},
		pending.Bootstrapper{},
		&ipfs.Bootstrapper{},
		&nftv3.Bootstrapper{},
		&p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&entity.Bootstrapper{},
		httpv2.Bootstrapper{},
		&httpv3.Bootstrapper{},
	}
}

// PopulateCommandBootstrappers adds all the bootstrapper implementations relevant for one off commands
func (m *MainBootstrapper) PopulateCommandBootstrappers() {
	m.Bootstrappers = []bootstrap.Bootstrapper{
		&version.Bootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&jobs.Bootstrapper{},
		centchain.Bootstrapper{},
		&configstore.Bootstrapper{},
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
			log.Errorf("Error encountered while bootstrapping %#v: %s", b, err)
			return err
		}
	}
	return nil
}
