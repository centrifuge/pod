package entityrelationship

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/anchors"
	"github.com/centrifuge/go-centrifuge/storage"
)

const (
	// BootstrappedEntityRelationshipService maps to the service for entity relationships
	BootstrappedEntityRelationshipService string = "BootstrappedEntityRelationshipService"
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

	db, ok := ctx[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return errors.New("storage repository not initialised")
	}

	repo, ok := ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	if !ok {
		return errors.New("document db repository not initialised")
	}

	entityRepo := newDBRepository(db, repo)
	repo.Register(&EntityRelationship{})

	identityService, ok := ctx[v2.BootstrappedIdentityServiceV2].(v2.Service)
	if !ok {
		return errors.New("identity service v2 not initialised")
	}

	anchorSrv, ok := ctx[pallets.BootstrappedAnchorService].(anchors.API)
	if !ok {
		return errors.New("anchor service not initialised")
	}

	// register service
	srv := NewService(
		docSrv,
		entityRepo,
		anchorSrv,
		identityService,
	)

	err := registry.Register(documenttypes.EntityRelationshipDataTypeUrl, srv)
	if err != nil {
		return errors.New("failed to register entity relationship service: %v", err)
	}

	err = registry.Register(Scheme, srv)
	if err != nil {
		return errors.New("failed to register entity relationship service: %v", err)
	}

	ctx[BootstrappedEntityRelationshipService] = srv

	return nil
}
