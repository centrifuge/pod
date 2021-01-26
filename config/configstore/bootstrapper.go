package configstore

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/storage"
)

// Bootstrapper implements bootstrap.Bootstrapper to initialise configstore package.
type Bootstrapper struct{}

// Bootstrap takes the passed in config file, loads the config and puts the config back into context.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, ok := context[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("could not find the bootstrapped config"))
	}
	configdb, ok := context[storage.BootstrappedConfigDB].(storage.Repository)
	if !ok {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("could not find the storage repository"))
	}
	idFactory, ok := context[identity.BootstrappedDIDFactory].(identity.Factory)
	if !ok {
		return errors.New("identity factory service not initialised")
	}
	idService, ok := context[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	dispatcher, ok := context[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	if !ok {
		return errors.New("dispatcher not initialised")
	}

	ethClient, ok := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	if !ok {
		return errors.New("ethereum client not initialised")
	}

	repo := &repo{configdb}
	// TODO(ved):  add idFactoryV2 instance
	service := &service{
		repo:      repo,
		idFactory: idFactory,
		idService: idService,
		protocolSetterFinder: func() ProtocolSetter {
			return context[bootstrap.BootstrappedPeer].(ProtocolSetter)
		},
		dispatcher: dispatcher,
	}

	// TODO(ved):  add idFactoryV2 instance and registration
	go dispatcher.RegisterRunner(generateIdentityRunnerName, generateIdentityRunner{
		idFactory: nil,
		ethClient: ethClient,
		repo:      repo,
	})

	// install the file based config every time so that file updates are reflected in the db, direct updates to db are not allowed
	nc := NewNodeConfig(cfg)
	configdb.Register(nc)
	nc, err := service.CreateConfig(nc)
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}

	tc, err := NewAccount(nc.GetEthereumDefaultAccountName(), cfg)
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}
	configdb.Register(tc)
	i, err := nc.GetIdentityID()
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}
	_, err = service.GetAccount(i)
	// if main account doesn't exist in the db, add it
	// Another additional check we can do is check if there are more than 0 accounts in the db but main account is not, then it might indicate a problem
	if err != nil {
		_, err = service.CreateAccount(tc)
		if err != nil {
			return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
		}
	}
	context[config.BootstrappedConfigStorage] = service
	return nil
}
