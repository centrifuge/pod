package purchaseorder

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
)

const (
	// BootstrappedPOHandler maps to grc handler for PO
	BootstrappedPOHandler = "BootstrappedPOHandler"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises required services for purchaseorder.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("service registry not initialised")
	}

	repo, ok := ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	if !ok {
		return errors.New("document db repository not initialised")
	}
	repo.Register(&PurchaseOrder{})

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue server not initialised")
	}

	txService, ok := ctx[transactions.BootstrappedService].(transactions.Manager)
	if !ok {
		return errors.New("transaction service not initialised")
	}

	cfgSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config service not initialised")
	}

	// register service
	srv := DefaultService(docSrv, repo, queueSrv, txService)
	err := registry.Register(documenttypes.PurchaseOrderDataTypeUrl, srv)
	if err != nil {
		return errors.New("failed to register purchase order service")
	}

	ctx[BootstrappedPOHandler] = GRPCHandler(cfgSrv, srv)

	return nil
}
