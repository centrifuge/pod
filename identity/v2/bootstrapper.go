package v2

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	BootstrappedIdentityServiceV2 = "BootstrappedIdentityServiceV2"
)

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
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

	proxyAPI, ok := context[pallets.BootstrappedProxyAPI].(proxy.API)

	if !ok {
		return errors.New("proxy API not initialised")
	}

	keystoreAPI, ok := context[pallets.BootstrappedKeystoreAPI].(keystore.API)

	if !ok {
		return errors.New("keystore API not initialised")
	}

	identityServiceV2 := NewService(cfgService, centAPI, keystoreAPI, proxyAPI, protocolIDDispatcher)

	context[BootstrappedIdentityServiceV2] = identityServiceV2

	return nil
}
