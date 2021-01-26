package anchors

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
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

	if _, ok := ctx[centchain.BootstrappedCentChainClient]; !ok {
		return errors.New("centchain client hasn't been initialized")
	}
	client := ctx[centchain.BootstrappedCentChainClient].(centchain.API)

	jobsMan, ok := ctx[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("jobs repository not initialised")
	}

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue hasn't been initialized")
	}

	repo := NewRepository(client, jobsMan)
	srv := newService(cfg, repo, queueSrv, jobsMan)
	ctx[BootstrappedAnchorService] = srv
	return nil
}
