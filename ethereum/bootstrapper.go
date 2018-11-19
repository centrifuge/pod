package ethereum

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

type Bootstrapper struct{}

func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[bootstrap.BootstrappedConfig].(*config.Configuration)
	client, err := NewGethClient(cfg)
	if err != nil {
		return err
	}
	SetClient(client)
	context[bootstrap.BootstrappedEthereumClient] = client
	return nil
}
