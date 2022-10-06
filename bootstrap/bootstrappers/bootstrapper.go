package bootstrappers

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/http"
	httpv2 "github.com/centrifuge/go-centrifuge/http/v2"
	httpv3 "github.com/centrifuge/go-centrifuge/http/v3"
	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/version"
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
		jobs.Bootstrapper{},
		centchain.Bootstrapper{},
		&configstore.Bootstrapper{},
		&pallets.Bootstrapper{},
		&dispatcher.Bootstrapper{},
		&identityv2.Bootstrapper{},
		&anchors.Bootstrapper{},
		documents.Bootstrapper{},
		http.Bootstrapper{},
		&entityrelationship.Bootstrapper{},
		generic.Bootstrapper{},
		pending.Bootstrapper{},
		&ipfs_pinning.Bootstrapper{},
		&nftv3.Bootstrapper{},
		p2p.Bootstrapper{},
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
		jobs.Bootstrapper{},
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
