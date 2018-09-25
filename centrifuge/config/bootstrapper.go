package config

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrap"

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	Config.InitializeViper()
	context[bootstrap.BootstrappedConfig] = Config
	return nil
}
