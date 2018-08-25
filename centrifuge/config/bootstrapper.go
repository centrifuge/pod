package config

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	Config.InitializeViper()
	context[bootstrapper.BOOTSTRAPPED_CONFIG] = Config
	return nil
}
