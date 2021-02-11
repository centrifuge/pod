package entity

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

const (
	// BootstrappedEntityService maps to the service for entities
	BootstrappedEntityService string = "BootstrappedEntityService"
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

	factory, ok := ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	if !ok {
		return errors.New("identity factory not initialised")
	}

	erService, ok := ctx[entityrelationship.BootstrappedEntityRelationshipService].(entityrelationship.Service)
	if !ok {
		return errors.New("entity relation service not initialised")
	}

	processor, ok := ctx[documents.BootstrappedAnchorProcessor].(documents.DocumentRequestProcessor)
	if !ok {
		return errors.New("processor not initialised")
	}

	anchorSrv, ok := ctx[anchors.BootstrappedAnchorService].(anchors.Service)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	didService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialized")
	}

	// register service
	srv := DefaultService(
		docSrv,
		repo,
		factory, erService, anchorSrv, processor, func() documents.ValidatorGroup {
			return documents.PostAnchoredValidator(didService, anchorSrv)
		})

	err := registry.Register(documenttypes.EntityDataTypeUrl, srv)
	if err != nil {
		return errors.New("failed to register entity service: %v", err)
	}

	err = registry.Register(Scheme, srv)
	if err != nil {
		return errors.New("failed to register entity service: %v", err)
	}

	ctx[BootstrappedEntityService] = srv

	return nil
}
