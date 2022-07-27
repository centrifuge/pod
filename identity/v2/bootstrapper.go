package v2

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/jobs"
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
	keystoreAPI := keystore.NewAPI(centAPI)

	context[BootstrappedKeystoreAPI] = keystoreAPI

	proxyAPI := proxy.NewAPI(centAPI)

	context[BootstrappedProxyAPI] = proxyAPI

	dispatcher, ok := context[jobs.BootstrappedDispatcher].(jobs.Dispatcher)

	if !ok {
		return errors.New("dispatcher not initialised")
	}

	go dispatcher.RegisterRunner(addKeysJob, &AddKeysJob{
		keystoreAPI: keystoreAPI,
	})

	configService, ok := context[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config storage not initialised")
	}

	identityServiceV2 := NewService(configService, centAPI, dispatcher, keystoreAPI)

	context[BootstrappedIdentityServiceV2] = identityServiceV2

	return nil
}
