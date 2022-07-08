package ideth

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the factory contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	//// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	//cfg, err := config.RetrieveConfig(false, context)
	//if err != nil {
	//	return err
	//}
	//
	//client := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	//
	//dispatcher := context[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	//factory := factroy{
	//	factoryAddress:  factoryAddress,
	//	factoryContract: factoryContract,
	//	client:          client,
	//	config:          cfg,
	//}
	//context[identity.BootstrappedDIDFactory] = factory
	//service := NewService(client, dispatcher, cfg)
	//context[identity.BootstrappedDIDService] = service
	return nil
}
