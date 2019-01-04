package ethereum

import "github.com/centrifuge/go-centrifuge/config/configstore"

// BootstrappedEthereumClient is a key to mapped client in bootstrap context.
const BootstrappedEthereumClient string = "BootstrappedEthereumClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises ethereum client.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(false, context)
	if err != nil {
		return err
	}
	client, err := NewGethClient(cfg)
	if err != nil {
		return err
	}
	SetClient(client)
	context[BootstrappedEthereumClient] = client
	return nil
}
