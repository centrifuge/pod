package coredocument

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// Bootstrapper to initialise processor
type Bootstrapper struct{}

func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("config hasn't been initialized")
	}

	anchorRepo, ok := ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	idService, ok := ctx[identity.BootstrappedIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	p2pClient, ok := ctx[bootstrap.BootstrappedP2PClient].(client)
	if !ok {
		return errors.New("p2p client not initialised")
	}

	ctx[bootstrap.BootstrappedCoreDocProc] = DefaultProcessor(idService, p2pClient, anchorRepo, cfg)
	return nil
}
