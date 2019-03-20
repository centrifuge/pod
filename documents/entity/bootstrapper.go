package entity

/*
const (
	// BootstrappedEntityHandler maps to grpc handler for entities
	BootstrappedEntityHandler string = "BootstrappedEntityHandler"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("service registry not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	repo, ok := ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	if !ok {
		return errors.New("document db repository not initialised")
	}
	repo.Register(&Entity{})

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue server not initialised")
	}

	txManager, ok := ctx[transactions.BootstrappedService].(transactions.Manager)
	if !ok {
		return errors.New("transaction service not initialised")
	}

	cfgSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config service not initialised")
	}

	// register service
	srv := DefaultService(
		docSrv,
		repo,
		queueSrv, txManager)

	err := registry.Register(documenttypes.EntityDataTypeUrl, srv)
	if err != nil {
		return errors.New("failed to register entity service: %v", err)
	}

	ctx[BootstrappedEntityHandler] = GRPCHandler(cfgSrv, srv)
	return nil
}
*/
