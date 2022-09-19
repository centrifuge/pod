package v2

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	dispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	BootstrappedIdentityServiceV2 = "BootstrappedIdentityServiceV2"
)

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	return b.bootstrap(context)
}

func (b *Bootstrapper) bootstrap(context map[string]interface{}) error {
	centAPI, ok := context[centchain.BootstrappedCentChainClient].(centchain.API)

	if !ok {
		return errors.New("centchain API not initialised")
	}

	cfgService, ok := context[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config service not initialised")
	}

	protocolIDDispatcher, ok := context[dispatcher.BootstrappedProtocolIDDispatcher].(dispatcher.Dispatcher[protocol.ID])

	if !ok {
		return errors.New("protocol ID dispatcher not initialised")
	}

	keystoreAPI := context[pallets.BootstrappedKeystoreAPI].(keystore.API)

	identityServiceV2 := NewService(cfgService, centAPI, keystoreAPI, protocolIDDispatcher)

	context[BootstrappedIdentityServiceV2] = identityServiceV2

	return nil
}
