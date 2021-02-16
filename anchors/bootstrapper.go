package anchors

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
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

	client := ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	srv := newService(cfg, client)
	ctx[BootstrappedAnchorService] = srv
	return nil
}
