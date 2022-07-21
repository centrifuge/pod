package entityrelationship

import (
	"context"

	v2 "github.com/centrifuge/go-centrifuge/identity/v2"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// GetEntityRelationships returns a list of the latest versions of the relevant entity relationship based on an entity id
	GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Document, error)
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo            repository
	anchorSrv       anchors.Service
	identityService v2.Service
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo repository,
	anchorSrv anchors.Service,
	identityService v2.Service,
) Service {
	return service{
		repo:            repo,
		Service:         srv,
		anchorSrv:       anchorSrv,
		identityService: identityService,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an entity
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Document, error) {
	er := new(EntityRelationship)
	err := er.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return er, nil
}

// GetEntityRelationships returns the latest versions of the entity relationships that involve the entityID passed in
func (s service) GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Document, error) {
	var relationships []documents.Document
	if entityID == nil {
		return nil, documents.ErrPayloadNil
	}

	selfIdentity, err := contextutil.Identity(ctx)
	if err != nil {
		return nil, errors.New("failed to get self ID")
	}

	relevant, err := s.repo.ListAllRelationships(entityID, selfIdentity)
	if err != nil {
		return nil, err
	}

	for _, v := range relevant {
		r, err := s.GetCurrentVersion(ctx, v)
		if err != nil {
			return nil, err
		}
		tokens, err := r.GetAccessTokens()
		if err != nil {
			return nil, err
		}

		if len(tokens) < 1 {
			continue
		}

		relationships = append(relationships, r)
	}

	return relationships, nil
}

// New returns a new uninitialised EntityRelationship.
func (s service) New(_ string) (documents.Document, error) {
	return new(EntityRelationship), nil
}

// Validate takes care of document validation
func (s service) Validate(ctx context.Context, model documents.Document, old documents.Document) error {
	return fieldValidator(s.identityService).Validate(old, model)
}
