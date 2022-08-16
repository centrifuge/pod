package configstore

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
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

	repo := NewDBRepository(configdb)
	service := NewService(repo)

	acc := &Account{}

	configdb.Register(acc)

	nodeCfg := config.NewNodeConfig(cfg)
	configdb.Register(nodeCfg)
	err := service.CreateConfig(nodeCfg)
	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("%v", err))
	}

	if err := processNodeAdmin(cfg, service, configdb); err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, errors.New("couldn't process node admin: %s", err))
	}

	context[config.BootstrappedConfigStorage] = service
	return nil
}

func processNodeAdmin(cfg config.Configuration, cfgService config.Service, configDB storage.Repository) error {
	nodeAdmin, err := getNodeAdmin(cfg)

	if err != nil {
		return err
	}

	storedAdmin, err := cfgService.GetNodeAdmin()

	if err != nil {
		configDB.Register(nodeAdmin)

		return cfgService.CreateNodeAdmin(nodeAdmin)
	}

	if !storedAdmin.GetAccountID().Equal(nodeAdmin.GetAccountID()) {
		return cfgService.UpdateNodeAdmin(nodeAdmin)
	}

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
