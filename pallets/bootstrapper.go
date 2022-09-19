package pallets

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
)

const (
	BootstrappedKeystoreAPI = "BootstrappedKeystoreAPI"
	BootstrappedProxyAPI    = "BootstrappedProxyAPI"
	BootstrappedUniquesAPI  = "BootstrappedUniquesAPI"
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

	proxyAPI := proxy.NewAPI(centAPI)

	context[BootstrappedProxyAPI] = proxyAPI

	keystoreAPI := keystore.NewAPI(cfgService, centAPI, proxyAPI)

	context[BootstrappedKeystoreAPI] = keystoreAPI

	uniquesAPI := uniques.NewAPI(cfgService, centAPI, proxyAPI)

	context[BootstrappedUniquesAPI] = uniquesAPI

	return nil
}
