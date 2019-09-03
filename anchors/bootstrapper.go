package anchors

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
)

const (
	// BootstrappedAnchorRepo is used as a key to map the configured anchor repository through context.
	BootstrappedAnchorRepo string = "BootstrappedAnchorRepo"

	// ErrAnchorRepoNotInitialised is a sentinel error when repository is not initialised
	ErrAnchorRepoNotInitialised = errors.Error("anchor repository not initialised")
)

// Bootstrapper implements bootstrapper.Bootstrapper for package requirement initialisations.
type Bootstrapper struct{}

// Bootstrap initializes the anchorRepositoryContract as well as the anchorConfirmationTask that depends on it.
// the anchorConfirmationTask is added to be registered on the Queue at queue.Bootstrapper.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(false, ctx)
	if err != nil {
		return err
	}

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	anchorContractAddr := cfg.GetContractAddress(config.AnchorRepo)

	repositoryContract, err := NewAnchorContract(anchorContractAddr, client.GetEthClient())
	if err != nil {
		return err
	}

	jobsMan, ok := ctx[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("jobs repository not initialised")
	}

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue hasn't been initialized")
	}

	repo := newService(cfg, repositoryContract, queueSrv, client, jobsMan)
	ctx[BootstrappedAnchorRepo] = repo

	return nil
}
