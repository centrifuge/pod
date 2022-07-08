package configstore

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
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

	idFactoryV2, ok := context[identity.BootstrappedDIDFactory].(identity.Factory)
	if !ok {
		return errors.New("configstore: identity factory not initialised")
	}

	idService, ok := context[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	dispatcher, ok := context[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	if !ok {
		return errors.New("dispatcher not initialised")
	}

	repo := &repo{configdb}
	service := &service{
		repo:      repo,
		idFactory: idFactory,
		idService: idService,
		protocolSetterFinder: func() ProtocolSetter {
			return context[bootstrap.BootstrappedPeer].(ProtocolSetter)
		},
		dispatcher:  dispatcher,
		idFactoryV2: idFactoryV2,
	}

	configdb.Register(cfg)
	_, err := service.CreateConfig(cfg)
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}

	context[config.BootstrappedConfigStorage] = service
	return nil
}
