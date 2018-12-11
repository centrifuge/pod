package config

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/storage"
)

// Bootstrap constants are keys to the value mappings in context bootstrap.
const (
	BootstrappedConfig        string = "BootstrappedConfig"
	BootstrappedConfigFile    string = "BootstrappedConfigFile"
	BootstrappedLevelDB       string = "BootstrappedLevelDB"
	BootstrappedConfigLevelDB string = "BootstrappedConfigLevelDB"
)

// Bootstrapper implements bootstrap.Bootstrapper to initialise config package.
type Bootstrapper struct{}

// Bootstrap takes the passed in config file, loads the config and puts the config back into context.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[BootstrappedConfigFile]; !ok {
		return errors.New("config file hasn't been provided")
	}
	cfgFile := context[BootstrappedConfigFile].(string)
	cfg := LoadConfiguration(cfgFile)
	context[BootstrappedConfig] = cfg

	configLevelDB, err := storage.NewLevelDBStorage(cfg.GetConfigStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init config level db: %v", err)
	}
	context[BootstrappedConfigLevelDB] = configLevelDB

	levelDB, err := storage.NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}

	context[BootstrappedLevelDB] = levelDB
	return nil
}
