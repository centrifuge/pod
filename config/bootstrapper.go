package config

import (
	"errors"

)

const (
	BootstrappedConfig       string = "BootstrappedConfig"
	BootstrappedConfigFile   string = "BootstrappedConfigFile"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[BootstrappedConfigFile]; !ok {
		return errors.New("config file hasn't been provided")
	}
	cfgFile := context[BootstrappedConfigFile].(string)
	context[BootstrappedConfig] = NewConfiguration(cfgFile)
	return nil
}
