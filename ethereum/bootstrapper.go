package ethereum

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/config"
)

const BootstrappedEthereumClient string = "BootstrappedEthereumClient"

type Bootstrapper struct{}

func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(*config.Configuration)
	client, err := NewGethClient(cfg)
	if err != nil {
		return err
	}
	SetClient(client)
	context[BootstrappedEthereumClient] = client
	return nil
}
