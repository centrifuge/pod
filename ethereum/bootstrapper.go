package ethereum

// BootstrappedEthereumClient is a key to mapped client in bootstrap context.
const BootstrappedEthereumClient string = "BootstrappedEthereumClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises ethereum client.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	//cfg, err := config.RetrieveConfig(false, ctx)
	//if err != nil {
	//	return err
	//}
	//
	//client, err := NewGethClient(cfg)
	//if err != nil {
	//	return err
	//}
	//
	//SetClient(client)
	//ctx[BootstrappedEthereumClient] = client
	return nil
}
