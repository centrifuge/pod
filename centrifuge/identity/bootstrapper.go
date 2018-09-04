package identity

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the IdentityFactoryContract as well as the IdRegistrationConfirmationTask that depends on it.
// the IdRegistrationConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	identityContract, err := getIdentityFactoryContract()
	if err != nil {
		return err
	}
	return queue.InstallQueuedTask(context, createIdRegistrationConfirmationTask(identityContract))
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func createIdRegistrationConfirmationTask(identityContract *EthereumIdentityFactoryContract) queue.QueuedTask {
	return NewIdRegistrationConfirmationTask(
		&identityContract.EthereumIdentityFactoryContractFilterer,
		ethereum.DefaultWaitForTransactionMiningContext)
}
