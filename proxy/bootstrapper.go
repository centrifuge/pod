package proxy

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	// BootstrappedProxyService is used as a key to map the configured proxy service through context.
	BootstrappedProxyService string = "BootstrappedProxyService"

	// ErrProxyRepoNotInitialised is a sentinel error when repository is not initialised
	ErrProxyRepoNotInitialised = errors.Error("proxy repository not initialised")
)

// Bootstrapper implements bootstrapper.Bootstrapper for package requirement initialisations.
type Bootstrapper struct{}

// Bootstrap initializes the Proxy Service
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	client := ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	srv := newService(client)
	ctx[BootstrappedProxyService] = srv
	return nil
}
