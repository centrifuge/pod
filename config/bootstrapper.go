package config

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	Config().InitializeViper()
	context[bootstrap.BootstrappedConfig] = Config()
	return nil
}
