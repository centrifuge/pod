package v2

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	dispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	BootstrappedKeystoreAPI       = "BootstrappedKeystoreAPI"
	BootstrappedProxyAPI          = "BootstrappedProxyAPI"
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

	proxyAPI := proxy.NewAPI(centAPI)

	context[BootstrappedProxyAPI] = proxyAPI

	keystoreAPI := keystore.NewAPI(centAPI, proxyAPI)

	context[BootstrappedKeystoreAPI] = keystoreAPI

	jobsDispatcher, ok := context[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)

	if !ok {
		return errors.New("dispatcher not initialised")
	}

	go jobsDispatcher.RegisterRunner(addKeysJob, &AddKeysJob{
		keystoreAPI: keystoreAPI,
	})

	configService, ok := context[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config storage not initialised")
	}

	protocolIDDispatcher, ok := context[dispatcher.BootstrappedProtocolIDDispatcher].(dispatcher.Dispatcher[protocol.ID])

	if !ok {
		return errors.New("protocol ID dispatcher not initialised")
	}

	identityServiceV2 := NewService(configService, centAPI, jobsDispatcher, keystoreAPI, protocolIDDispatcher)

	context[BootstrappedIdentityServiceV2] = identityServiceV2

	return nil
}
