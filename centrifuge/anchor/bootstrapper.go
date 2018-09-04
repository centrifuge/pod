package anchor

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the AnchorRegistryContract as well as the AnchoringConfirmationTask that depends on it.
// the AnchoringConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	anchorContract, err := getAnchorContract()
	if err != nil {
		return err
	}
	return queue.InstallQueuedTask(context, NewAnchoringConfirmationTask(&anchorContract.EthereumAnchorRegistryContractFilterer, ethereum.DefaultWaitForTransactionMiningContext))
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}
