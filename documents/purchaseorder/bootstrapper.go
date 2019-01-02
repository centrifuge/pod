package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/ethid"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises required services for purchaseorder.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(true, ctx)
	if err != nil {
		return err
	}

	p2pClient, ok := ctx[p2p.BootstrappedP2PClient].(p2p.Client)
	if !ok {
		return errors.New("p2p client not initialised")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("service registry not initialised")
	}

	anchorRepo, ok := ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	idService, ok := ctx[ethid.BootstrappedIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	repo, ok := ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	if !ok {
		return errors.New("document db repository not initialised")
	}
	repo.Register(&PurchaseOrder{})

	// register service
	srv := DefaultService(cfg, repo, coredocument.DefaultProcessor(idService, p2pClient, anchorRepo, cfg), anchorRepo, idService)
	err = registry.Register(documenttypes.PurchaseOrderDataTypeUrl, srv)
	if err != nil {
		return errors.New("failed to register purchase order service")
	}

	return nil
}
