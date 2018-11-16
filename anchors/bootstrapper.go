package anchors

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the AnchorRepositoryContract as well as the anchorConfirmationTask that depends on it.
// the anchorConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[bootstrap.BootstrappedConfig].(*config.Configuration)

	if _, ok := context[bootstrap.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	client := ethereum.GetClient()
	repositoryContract, err := NewEthereumAnchorRepositoryContract(cfg.GetContractAddress("anchorRepository"), client.GetEthClient())
	if err != nil {
		return err
	}

	anchorRepo := NewEthereumAnchorRepository(cfg, repositoryContract, ethereum.GetClient)
	setAnchorRepository(anchorRepo)
	if err != nil {
		return err
	}

	task := &anchorConfirmationTask{
		AnchorCommittedFilterer: &repositoryContract.EthereumAnchorRepositoryContractFilterer,
		EthContextInitializer:   ethereum.DefaultWaitForTransactionMiningContext,
	}

	return queue.InstallQueuedTask(context, task)
}
