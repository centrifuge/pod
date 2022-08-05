package config

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	logging "github.com/ipfs/go-log"
)

// Bootstrap constants are keys to the value mappings in context bootstrap.
const (
	// BootstrappedConfigFile points to the config file the node is bootstrapped with
	BootstrappedConfigFile string = "BootstrappedConfigFile"

	// BootstrappedConfigStorage indicates that config storage has been bootstrapped and its the key for config storage service in the bootstrap context
	BootstrappedConfigStorage string = "BootstrappedConfigStorage"
)

// Bootstrapper implements bootstrap.Bootstrapper to initialise config package.
type Bootstrapper struct{}

// Bootstrap takes the passed in config file, loads the config and puts the config back into context.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[BootstrappedConfigFile]; !ok {
		return ErrConfigFileBootstrapNotFound
	}
	cfgFile := context[BootstrappedConfigFile].(string)
	c := LoadConfiguration(cfgFile)
	context[bootstrap.BootstrappedConfig] = c
	if c.IsDebugLogEnabled() {
		logging.SetAllLoggers(logging.LevelDebug)
	}
	return nil
}

const (
	defaultConfigFile = "build/configs/testing_config.yaml"
)

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; ok {
		return nil
	}

	cfgFile, ok := context[BootstrappedConfigFile].(string)

	// Load the file that's provided in the context or the default.
	if !ok {
		cfgFile = defaultConfigFile
	}

	context[bootstrap.BootstrappedConfig] = LoadConfiguration(cfgFile)

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}
