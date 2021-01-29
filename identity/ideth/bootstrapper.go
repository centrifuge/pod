package ideth

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/common"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the factory contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	cfg, err := config.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

	client := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	factoryAddress := getFactoryAddress(cfg)
	factoryContract, err := bindFactory(factoryAddress, client)
	if err != nil {
		return err
	}
	jobManager := context[jobs.BootstrappedService].(jobs.Manager)
	dispatcher := context[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	queueSrv := context[bootstrap.BootstrappedQueueServer].(*queue.Server)
	factory := factroy{
		factoryAddress:  factoryAddress,
		factoryContract: factoryContract,
		client:          client,
		config:          cfg,
	}
	context[identity.BootstrappedDIDFactory] = factory
	service := NewService(client, dispatcher, jobManager, queueSrv, cfg)
	context[identity.BootstrappedDIDService] = service
	return nil
}

func bindFactory(factoryAddress common.Address, client ethereum.Client) (*FactoryContract, error) {
	return NewFactoryContract(factoryAddress, client.GetEthClient())
}

func getFactoryAddress(cfg config.Configuration) common.Address {
	return cfg.GetContractAddress(config.IdentityFactory)
}
