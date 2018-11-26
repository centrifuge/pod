package ethereum

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/config"
)

// BootstrappedEthereumClient is a key to mapped client in bootstrap context.
const BootstrappedEthereumClient string = "BootstrappedEthereumClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises ethereum client.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(Config)
	client, err := NewGethClient(cfg)
	if err != nil {
		return err
	}
	SetClient(client)
	context[BootstrappedEthereumClient] = client
	return nil
}
