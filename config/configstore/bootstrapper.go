package configstore

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
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

	dispatcher, ok := context[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	if !ok {
		return errors.New("dispatcher not initialised")
	}

	repo := &repo{configdb}
	service := NewService(
		repo,
		dispatcher,
		func() ProtocolSetter {
			return context[bootstrap.BootstrappedPeer].(ProtocolSetter)
		},
	)

	acc := &Account{}

	configdb.Register(acc)

	nodeCfg := config.NewNodeConfig(cfg)
	configdb.Register(nodeCfg)
	_, err := service.CreateConfig(nodeCfg)
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}

	nodeAdmin, err := getNodeAdmin(cfg)

	if err != nil {
		return err
	}

	configdb.Register(nodeAdmin)

	_, err = service.CreateNodeAdmin(nodeAdmin)

	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("couldn't create node admin: %s", err))
	}

	context[config.BootstrappedConfigStorage] = service
	return nil
}

func getNodeAdmin(cfg config.Configuration) (config.NodeAdmin, error) {
	adminPubKeyPath, _ := cfg.GetNodeAdminKeyPair()

	adminPubKey, err := utils.ReadKeyFromPemFile(adminPubKeyPath, utils.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't read admin public key: %w", err)
	}

	adminAccountID, err := types.NewAccountID(adminPubKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't create admin account ID: %w", err)
	}

	return NewNodeAdmin(adminAccountID), nil
}
