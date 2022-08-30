package configstore

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey"
	"github.com/vedhavyas/go-subkey/sr25519"
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

	repo.RegisterAccount(acc)

	nodeCfg := config.NewNodeConfig(cfg)

	repo.RegisterConfig(nodeCfg)

	if err := service.CreateConfig(nodeCfg); err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, fmt.Errorf("couldn't create config: %w", err))
	}

	nodeAdmin, err := getNodeAdmin(cfg)

	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, fmt.Errorf("couldn't get node admin: %w", err))
	}

	repo.RegisterNodeAdmin(nodeAdmin)

	if err := service.CreateNodeAdmin(nodeAdmin); err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, fmt.Errorf("couldn't create node admin: %w", err))
	}

	podOperator, err := getPodOperator(cfg)

	if err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, fmt.Errorf("couldn't get pod operator: %w", err))
	}

	repo.RegisterPodOperator(podOperator)

	if err := service.CreatePodOperator(podOperator); err != nil {
		return errors.NewTypedError(config.ErrConfigBootstrap, fmt.Errorf("couldn't create pod operator: %w", err))
	}

	context[config.BootstrappedConfigStorage] = service
	return nil
}

func getPodOperator(cfg config.Configuration) (config.PodOperator, error) {
	kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, cfg.GetPodOperatorSecretSeed())

	if err != nil {
		return nil, fmt.Errorf("couldn't derive pod operator key pair: %w", err)
	}

	accountID, err := types.NewAccountID(kp.AccountID())

	if err != nil {
		return nil, fmt.Errorf("couldn't create pod operator account ID: %w", err)
	}

	return NewPodOperator(cfg.GetPodOperatorSecretSeed(), accountID), nil
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
