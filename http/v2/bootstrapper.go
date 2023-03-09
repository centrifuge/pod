package v2

import (
	"errors"
	"fmt"

	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/documents/entity"
	"github.com/centrifuge/pod/documents/entityrelationship"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/pending"
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

	configService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)

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

	service, err := NewService(
		pendingDocSrv,
		dispatcher,
		configService,
		entitySrv,
		identityService,
		erSrv,
		docSrv,
	)

	if err != nil {
		return fmt.Errorf("couldn't create new service: %w", err)
	}

	ctx[BootstrappedService] = service

	return nil
}
