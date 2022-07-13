package v2

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
)

const (
	BootstrappedKeystoreAPI       = "BootstrappedKeystoreAPI"
	BootstrappedProxyAPI          = "BootstrappedProxyAPI"
	BootstrappedIdentityServiceV2 = "BootstrappedIdentityServiceV2"
)

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	centAPI := context[centchain.BootstrappedCentChainClient].(centchain.API)
	keystoreAPI := keystore.NewAPI(centAPI)

	context[BootstrappedKeystoreAPI] = keystoreAPI

	proxyAPI := proxy.NewAPI(centAPI)

	context[BootstrappedProxyAPI] = proxyAPI

	accountsSrv := context[config.BootstrappedConfigStorage].(config.Service)

	identityServiceV2 := NewService(accountsSrv, centAPI, keystoreAPI)

	context[BootstrappedIdentityServiceV2] = identityServiceV2

	return nil
}
