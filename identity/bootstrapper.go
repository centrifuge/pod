package identity

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Bootstrapper struct {
}

// Bootstrap initializes the IdentityFactoryContract as well as the idRegistrationConfirmationTask that depends on it.
// the idRegistrationConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[bootstrap.BootstrappedConfig].(*config.Configuration)

	if _, ok := context[bootstrap.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	idFactory, err := getIdentityFactoryContract(cfg.GetContractAddress("identityFactory"))
	if err != nil {
		return err
	}

	registryContract, err := getIdentityRegistryContract(cfg.GetContractAddress("identityRegistry"))
	if err != nil {
		return err
	}

	IDService = NewEthereumIdentityService(cfg, idFactory, registryContract, ethereum.GetClient,
		func(address common.Address, backend bind.ContractBackend) (contract, error) {
			return NewEthereumIdentityContract(address, backend)
		})

	err = queue.InstallQueuedTask(context,
		newIdRegistrationConfirmationTask(&idFactory.EthereumIdentityFactoryContractFilterer, ethereum.DefaultWaitForTransactionMiningContext))
	if err != nil {
		return err
	}

	err = queue.InstallQueuedTask(context,
		newKeyRegistrationConfirmationTask(ethereum.DefaultWaitForTransactionMiningContext, registryContract, cfg))
	if err != nil {
		return err
	}
	return nil
}

func getIdentityFactoryContract(factoryAddress common.Address) (identityFactoryContract *EthereumIdentityFactoryContract, err error) {
	client := ethereum.GetClient()
	return NewEthereumIdentityFactoryContract(factoryAddress, client.GetEthClient())
}

func getIdentityRegistryContract(registryAddress common.Address) (identityRegistryContract *EthereumIdentityRegistryContract, err error) {
	client := ethereum.GetClient()
	return NewEthereumIdentityRegistryContract(registryAddress, client.GetEthClient())
}
