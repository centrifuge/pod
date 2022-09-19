package anchors

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
)

const (
	// BootstrappedAnchorService is used as a key to map the configured anchor service through context.
	BootstrappedAnchorService string = "BootstrappedAnchorService"

	// ErrAnchorRepoNotInitialised is a sentinel error when repository is not initialised
	ErrAnchorRepoNotInitialised = errors.Error("anchor repository not initialised")
)

// Bootstrapper implements bootstrapper.Bootstrapper for package requirement initialisations.
type Bootstrapper struct{}

// Bootstrap initializes the anchorRepositoryContract as well as the anchorConfirmationTask that depends on it.
// the anchorConfirmationTask is added to be registered on the Queue at queue.Bootstrapper.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := config.RetrieveConfig(false, ctx)
	if err != nil {
		return err
	}

	client, ok := ctx[centchain.BootstrappedCentChainClient].(centchain.API)

	if !ok {
		return errors.New("cent chain API no initialised")
	}

	proxyAPI, ok := ctx[pallets.BootstrappedProxyAPI].(proxy.API)

	if !ok {
		return errors.New("proxy API no initialised")
	}

	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config storage not initialised")
	}

	srv := newService(cfg.GetCentChainAnchorLifespan(), cfgService, client, proxyAPI)
	ctx[BootstrappedAnchorService] = srv
	return nil
}
