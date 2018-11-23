package purchaseorder

import (
	"errors"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/storage"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises required services for purchaseorder.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := ctx[config.BootstrappedConfig].(*config.Configuration)

	if _, ok := ctx[storage.BootstrappedLevelDB]; !ok {
		return errors.New("could not initialize purchase order repository")
	}

	p2pClient, ok := ctx[p2p.BootstrappedP2PClient].(p2p.Client)
	if !ok {
		return fmt.Errorf("p2p client not initialised")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return fmt.Errorf("service registry not initialised")
	}

	anchorRepo, ok := ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	if !ok {
		return fmt.Errorf("anchor repository not initialised")
	}

	idService, ok := ctx[identity.BootstrappedIDService].(identity.Service)
	if !ok {
		return fmt.Errorf("identity service not initialised")
	}

	// register service
	srv := DefaultService(cfg, getRepository(), coredocument.DefaultProcessor(idService, p2pClient, anchorRepo, cfg), anchorRepo, idService)
	err := registry.Register(documenttypes.PurchaseOrderDataTypeUrl, srv)
	if err != nil {
		return fmt.Errorf("failed to register purchase order service")
	}

	return nil
}
