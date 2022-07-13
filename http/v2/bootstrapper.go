package v2

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pending"
)

// BootstrappedService key maps to the Service implementation in Bootstrap context.
const BootstrappedService = "V2 Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	pendingDocSrv := ctx[pending.BootstrappedPendingDocumentService].(pending.Service)
	accountsSrv := ctx[config.BootstrappedConfigStorage].(config.Service)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	entitySrv := ctx[entity.BootstrappedEntityService].(entity.Service)
	erSrv := ctx[entityrelationship.BootstrappedEntityRelationshipService].(entityrelationship.Service)
	docSrv := ctx[documents.BootstrappedDocumentService].(documents.Service)
	ctx[BootstrappedService] = Service{
		pendingDocSrv: pendingDocSrv,
		accountSrv:    accountsSrv,
		dispatcher:    dispatcher,
		entitySrv:     entitySrv,
		erSrv:         erSrv,
		docSrv:        docSrv,
	}
	return nil
}
