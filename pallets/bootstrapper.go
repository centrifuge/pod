package pallets

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/pallets/anchors"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
)

const (
	BootstrappedAnchorService = "BootstrappedAnchorService"
	BootstrappedKeystoreAPI   = "BootstrappedKeystoreAPI"
	BootstrappedProxyAPI      = "BootstrappedProxyAPI"
	BootstrappedUniquesAPI    = "BootstrappedUniquesAPI"
)

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, err := config.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

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

	anchorsAPI := anchors.NewAPI(cfg.GetCentChainAnchorLifespan(), cfgService, centAPI, proxyAPI)

	context[BootstrappedAnchorService] = anchorsAPI

	return nil
}
