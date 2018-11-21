package identity

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Bootstrapper struct {
}

// Bootstrap initializes the IdentityFactoryContract as well as the idRegistrationConfirmationTask that depends on it.
// the idRegistrationConfirmationTask is added to be registered on the queue at queue.Bootstrapper
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(Config)

	if _, ok := context[ethereum.BootstrappedEthereumClient]; !ok {
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

	if _, ok := context[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := context[bootstrap.BootstrappedQueueServer].(*queue.Server)

	IDService = NewEthereumIdentityService(cfg, idFactory, registryContract, queueSrv, ethereum.GetClient,
		func(address common.Address, backend bind.ContractBackend) (contract, error) {
			return NewEthereumIdentityContract(address, backend)
		})

	idRegTask := newIdRegistrationConfirmationTask(cfg.GetEthereumContextWaitTimeout(), &idFactory.EthereumIdentityFactoryContractFilterer, ethereum.DefaultWaitForTransactionMiningContext)
	keyRegTask := newKeyRegistrationConfirmationTask(ethereum.DefaultWaitForTransactionMiningContext, registryContract, cfg, queueSrv, ethereum.GetClient,
		func(address common.Address, backend bind.ContractBackend) (contract, error) {
			return NewEthereumIdentityContract(address, backend)
		})
	queueSrv.RegisterTaskType(idRegTask.TaskTypeName(), idRegTask)
	queueSrv.RegisterTaskType(keyRegTask.TaskTypeName(), keyRegTask)
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
