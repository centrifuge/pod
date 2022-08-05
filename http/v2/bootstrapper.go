package v2

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pending"
)

// BootstrappedService key maps to the Service implementation in Bootstrap context.
const BootstrappedService = "V2 Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	pendingDocSrv, ok := ctx[pending.BootstrappedPendingDocumentService].(pending.Service)
	if !ok {
		return errors.New("pending document service not initialised")
	}

	accountsSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config storage not initialised")
	}

	dispatcher, ok := ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)

	if !ok {
		return errors.New("job dispatcher not initialised")
	}

	entitySrv, ok := ctx[entity.BootstrappedEntityService].(entity.Service)

	if !ok {
		return errors.New("entity service not initialised")
	}

	erSrv, ok := ctx[entityrelationship.BootstrappedEntityRelationshipService].(entityrelationship.Service)

	if !ok {
		return errors.New("entity relationship service not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)

	if !ok {
		return errors.New("document service not initialised")
	}

	identityService, ok := ctx[v2.BootstrappedIdentityServiceV2].(v2.Service)

	if !ok {
		return errors.New("identity service not initialised")
	}

	ctx[BootstrappedService] = Service{
		pendingDocSrv:   pendingDocSrv,
		accountSrv:      accountsSrv,
		dispatcher:      dispatcher,
		entitySrv:       entitySrv,
		erSrv:           erSrv,
		docSrv:          docSrv,
		identityService: identityService,
	}

	return nil
}
