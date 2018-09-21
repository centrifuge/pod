package anchors

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrapper"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the AnchorRepositoryContract as well as the AnchoringConfirmationTask that depends on it.
// the AnchoringConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	repositoryContract, err := getRepositoryContract()
	if err != nil {
		return err
	}
	return queue.InstallQueuedTask(context, NewAnchoringConfirmationTask(&repositoryContract.EthereumAnchorRepositoryContractFilterer, ethereum.DefaultWaitForTransactionMiningContext))
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}
