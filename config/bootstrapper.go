package config

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfigFile]; !ok {
		return errors.New("config file hasn't been provided")
	}
	cfgFile := context[bootstrap.BootstrappedConfigFile].(string)
	context[bootstrap.BootstrappedConfig] = NewConfiguration(cfgFile)
	return nil
}
