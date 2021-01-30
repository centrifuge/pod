package p2p

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
)

// Bootstrapper implements Bootstrapper with p2p details
type Bootstrapper struct{}

// Bootstrap initiates p2p server and client into context
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := config.RetrieveConfig(true, ctx)
	if err != nil {
		return err
	}

	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("configstore not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	idService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	tokenRegistry, ok := ctx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	if !ok {
		return errors.New("token registry is not initialised")
	}

	ctx[bootstrap.BootstrappedPeer] = &peer{config: cfgService, idService: idService, handlerCreator: func() *receiver.Handler {
		return receiver.New(cfgService, receiver.HandshakeValidator(cfg.GetNetworkID(), idService), docSrv, tokenRegistry, idService)
	}}
	return nil
}
