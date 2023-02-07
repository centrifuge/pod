package pallets

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets/anchors"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
	"github.com/centrifuge/go-centrifuge/pallets/utility"
)

const (
	BootstrappedAnchorService = "BootstrappedAnchorService"
	BootstrappedKeystoreAPI   = "BootstrappedKeystoreAPI"
	BootstrappedProxyAPI      = "BootstrappedProxyAPI"
	BootstrappedUniquesAPI    = "BootstrappedUniquesAPI"
	BootstrappedUtilityAPI    = "BootstrappedUtilityAPI"
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

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return errors.ErrPodOperatorRetrieval
	}

	proxyAPI := proxy.NewAPI(centAPI)

	context[BootstrappedProxyAPI] = proxyAPI

	keystoreAPI := keystore.NewAPI(centAPI, proxyAPI, podOperator)

	context[BootstrappedKeystoreAPI] = keystoreAPI

	uniquesAPI := uniques.NewAPI(centAPI, proxyAPI, podOperator)

	context[BootstrappedUniquesAPI] = uniquesAPI

	anchorsAPI := anchors.NewAPI(centAPI, proxyAPI, cfg.GetCentChainAnchorLifespan(), podOperator)

	context[BootstrappedAnchorService] = anchorsAPI

	utilityAPI := utility.NewAPI(centAPI, proxyAPI, podOperator)

	context[BootstrappedUtilityAPI] = utilityAPI

	return nil
}
