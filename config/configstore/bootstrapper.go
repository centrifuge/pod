package configstore

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
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
	idService, ok := context[identity.BootstrappedIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	repo := &repo{configdb}
	service := &service{repo, idService, func() ProtocolSetter {
		return context[bootstrap.BootstrappedP2PServer].(ProtocolSetter)
	}}

	nc := NewNodeConfig(cfg)
	configdb.Register(nc)
	_, err := service.GetConfig()
	// if node config doesn't exist in the db, add it
	if err != nil {
		nc, err = service.CreateConfig(NewNodeConfig(cfg))
		if err != nil {
			return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
		}
	}
	tc, err := NewTenantConfig(nc.GetEthereumDefaultAccountName(), cfg)
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}
	configdb.Register(tc)
	i, err := nc.GetIdentityID()
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}
	_, err = service.GetTenant(i)
	// if main tenant config doesn't exist in the db, add it
	// Another additional check we can do is check if there are more than 0 tenant configs in the db but main tenant is not, then it might indicate a problem
	if err != nil {
		_, err = service.CreateTenant(tc)
		if err != nil {
			return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
		}
	}
	context[config.BootstrappedConfigStorage] = service
	return nil
}
