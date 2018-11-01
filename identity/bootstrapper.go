package identity

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the IdentityFactoryContract as well as the idRegistrationConfirmationTask that depends on it.
// the idRegistrationConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	if _, ok := context[bootstrap.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	idFactory, err := getIdentityFactoryContract()
	if err != nil {
		return err
	}

	registryContract, err := getIdentityRegistryContract()
	if err != nil {
		return err
	}

	IDService = NewEthereumIdentityService(config.Config, idFactory, registryContract)

	identityContract, err := getIdentityFactoryContract()
	if err != nil {
		return err
	}

	err = queue.InstallQueuedTask(context,
		newIdRegistrationConfirmationTask(&identityContract.EthereumIdentityFactoryContractFilterer, ethereum.DefaultWaitForTransactionMiningContext))
	if err != nil {
		return err
	}

	err = queue.InstallQueuedTask(context,
		newKeyRegistrationConfirmationTask(ethereum.DefaultWaitForTransactionMiningContext, registryContract, config.Config))
	if err != nil {
		return err
	}
	return nil
}

func getIdentityFactoryContract() (identityFactoryContract *EthereumIdentityFactoryContract, err error) {
	client := ethereum.GetClient()
	return NewEthereumIdentityFactoryContract(config.Config.GetContractAddress("identityFactory"), client.GetEthClient())
}

func getIdentityRegistryContract() (identityRegistryContract *EthereumIdentityRegistryContract, err error) {
	client := ethereum.GetClient()
	return NewEthereumIdentityRegistryContract(config.Config.GetContractAddress("identityRegistry"), client.GetEthClient())
}
