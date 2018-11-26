package anchors

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
)

// BootstrappedAnchorRepo is used as a key to map the configured anchor repository through context.
const BootstrappedAnchorRepo string = "BootstrappedAnchorRepo"

// Bootstrapper implements bootstrapper.Bootstrapper for package requirement initialisations.
type Bootstrapper struct{}

// Bootstrap initializes the anchorRepositoryContract as well as the anchorConfirmationTask that depends on it.
// the anchorConfirmationTask is added to be registered on the Queue at queue.Bootstrapper.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := ctx[config.BootstrappedConfig].(Config)

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	repositoryContract, err := NewEthereumAnchorRepositoryContract(cfg.GetContractAddress("anchorRepository"), client.GetEthClient())
	if err != nil {
		return err
	}

	if _, ok := ctx[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue server hasn't been initialized")
	}

	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	repo := newEthereumAnchorRepository(cfg, repositoryContract, queueSrv, ethereum.GetClient)
	ctx[BootstrappedAnchorRepo] = repo

	task := &anchorConfirmationTask{
		// Passing timeout as a common property for every request, if we need more fine-grain control per request then we will override by invoker
		Timeout:                 cfg.GetEthereumContextWaitTimeout(),
		AnchorCommittedFilterer: &repositoryContract.EthereumAnchorRepositoryContractFilterer,
		EthContextInitializer:   ethereum.DefaultWaitForTransactionMiningContext,
	}

	queueSrv.RegisterTaskType(task.TaskTypeName(), task)
	return nil
}
