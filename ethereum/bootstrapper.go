package ethereum

import (
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

// BootstrappedEthereumClient is a key to mapped client in bootstrap context.
const BootstrappedEthereumClient string = "BootstrappedEthereumClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises ethereum client.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[bootstrap.BootstrappedConfig].(Config)
	client, err := NewGethClient(cfg)
	if err != nil {
		return err
	}
	SetClient(client)
	context[BootstrappedEthereumClient] = client
	return nil
}
