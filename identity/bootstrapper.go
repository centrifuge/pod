package identity

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// BootstrappedIDService is used as a key to map the configured ID Service through context.
const BootstrappedIDService string = "BootstrappedIDService"

type Bootstrapper struct {
}

// Bootstrap initializes the IdentityFactoryContract as well as the idRegistrationConfirmationTask that depends on it.
// the idRegistrationConfirmationTask is added to be registered on the Queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(Config)

	if _, ok := context[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}
	gethClient := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	idFactory, err := getIdentityFactoryContract(cfg.GetContractAddress("identityFactory"), gethClient)
	if err != nil {
		return err
	}

	registryContract, err := getIdentityRegistryContract(cfg.GetContractAddress("identityRegistry"), gethClient)
	if err != nil {
		return err
	}

	context[BootstrappedIDService] = NewEthereumIdentityService(cfg, idFactory, registryContract, ethereum.GetClient,
		func(address common.Address, backend bind.ContractBackend) (contract, error) {
			return NewEthereumIdentityContract(address, backend)
		})

	err = queue.InstallQueuedTask(context,
		newIdRegistrationConfirmationTask(cfg.GetEthereumContextWaitTimeout(), &idFactory.EthereumIdentityFactoryContractFilterer, ethereum.DefaultWaitForTransactionMiningContext))
	if err != nil {
		return err
	}

	err = queue.InstallQueuedTask(context,
		newKeyRegistrationConfirmationTask(ethereum.DefaultWaitForTransactionMiningContext, registryContract, cfg, ethereum.GetClient,
			func(address common.Address, backend bind.ContractBackend) (contract, error) {
				return NewEthereumIdentityContract(address, backend)
			}))
	if err != nil {
		return err
	}
	return nil
}

func getIdentityFactoryContract(factoryAddress common.Address, ethClient ethereum.Client) (identityFactoryContract *EthereumIdentityFactoryContract, err error) {
	return NewEthereumIdentityFactoryContract(factoryAddress, ethClient.GetEthClient())
}

func getIdentityRegistryContract(registryAddress common.Address, ethClient ethereum.Client) (identityRegistryContract *EthereumIdentityRegistryContract, err error) {
	return NewEthereumIdentityRegistryContract(registryAddress, ethClient.GetEthClient())
}
