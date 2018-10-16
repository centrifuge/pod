package ethereum

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	client, err := NewClientConnection(config.Config)
	if err != nil {
		return err
	}
	SetConnection(client)
	context[bootstrap.BootstrappedEthereumClient] = client
	return nil
}
