package ideth

import (
	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the factory contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	cfg, err := configstore.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

	if _, ok := context[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}
	client := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	factoryAddress := getFactoryAddress(cfg)

	factoryContract, err := bindFactory(factoryAddress, client)
	if err != nil {
		return err
	}

	txManager, ok := context[transactions.BootstrappedService].(transactions.Manager)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	queueSrv, ok := context[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue hasn't been initialized")
	}

	factory := NewFactory(factoryContract, client, txManager, queueSrv, factoryAddress)
	context[identity.BootstrappedDIDFactory] = factory

	service := NewService(client, txManager, queueSrv)
	context[identity.BootstrappedDIDService] = service

	return nil
}

func bindFactory(factoryAddress common.Address, client ethereum.Client) (*FactoryContract, error) {
	return NewFactoryContract(factoryAddress, client.GetEthClient())
}

func getFactoryAddress(cfg config.Configuration) common.Address {
	return cfg.GetContractAddress(config.IdentityFactory)
}
