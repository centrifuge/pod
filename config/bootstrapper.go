package config

import (
	"errors"
)

// Bootstrap constants are keys to the value mappings in context bootstrap.
const (
	BootstrappedConfig     string = "BootstrappedConfig"
	BootstrappedConfigFile string = "BootstrappedConfigFile"
)

// Bootstrapper implements bootstrap.Bootstrapper to initialise config package.
type Bootstrapper struct{}

// Bootstrap takes the passed in config file, loads the config and puts the config back into context.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[BootstrappedConfigFile]; !ok {
		return errors.New("config file hasn't been provided")
	}
	cfgFile := context[BootstrappedConfigFile].(string)
	context[BootstrappedConfig] = LoadConfiguration(cfgFile)
	return nil
}
