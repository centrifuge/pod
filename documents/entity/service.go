package entity

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// GetEntityByRelationship returns the entity model from database or requests from granter
	GetEntityByRelationship(ctx context.Context, relationshipIdentifier []byte) (documents.Document, error)
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo                    documents.Repository
	factory                 identity.Factory
	processor               documents.DocumentRequestProcessor
	erService               entityrelationship.Service
	anchorSrv               anchors.Service
	receivedEntityValidator func() documents.ValidatorGroup
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	factory identity.Factory,
	erService entityrelationship.Service,
	anchorSrv anchors.Service,
	processor documents.DocumentRequestProcessor,
	receivedEntityValidator func() documents.ValidatorGroup,
) Service {
	return service{
		repo:                    repo,
		Service:                 srv,
		factory:                 factory,
		erService:               erService,
		anchorSrv:               anchorSrv,
		processor:               processor,
		receivedEntityValidator: receivedEntityValidator,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an entity
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Document, error) {
	entity := new(Entity)
	err := entity.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return entity, nil
}

// validateAndPersist validates the document, calculates the data root, and persists to DB
func (s service) validateAndPersist(ctx context.Context, old, new documents.Document, validator documents.Validator) (documents.Document, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	entity, ok := new.(*Entity)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// validate the entity
	err = validator.Validate(old, entity)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(selfDID[:], entity.CurrentVersion(), entity)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return entity, nil
}

// GetEntityByRelationship returns the entity model from database or requests from a granter peer
func (s service) GetEntityByRelationship(ctx context.Context, relationshipIdentifier []byte) (documents.Document, error) {
	model, err := s.erService.GetCurrentVersion(ctx, relationshipIdentifier)
	if err != nil {
		return nil, entityrelationship.ErrERNotFound
	}

	relationship, ok := model.(*entityrelationship.EntityRelationship)
	if !ok {
		return nil, entityrelationship.ErrNotEntityRelationship
	}

	return s.requestEntityWithRelationship(ctx, relationship)
}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Document, error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	entity, err := s.Service.GetCurrentVersion(ctx, documentID)
	if err != nil {
		return nil, documents.ErrDocumentNotFound
	}

	isCollaborator, err := entity.IsDIDCollaborator(did)
	if err != nil || !isCollaborator {
		return nil, documents.ErrDocumentNotFound
	}

	return entity, nil
}

func (s service) requestEntityWithRelationship(ctx context.Context, relationship *entityrelationship.EntityRelationship) (documents.Document, error) {
	accessTokens, err := relationship.GetAccessTokens()
	if err != nil {
		return nil, documents.ErrCDAttribute
	}

	// only one access token per entity relationship
	if len(accessTokens) != 1 {
		return nil, entityrelationship.ErrERNoToken
	}

	at := accessTokens[0]
	if !utils.IsSameByteSlice(at.DocumentIdentifier, relationship.Data.EntityIdentifier) {
		return nil, entityrelationship.ErrERInvalidIdentifier
	}

	granterDID, err := identity.NewDIDFromBytes(at.Granter)
	if err != nil {
		return nil, err
	}
	response, err := s.processor.RequestDocumentWithAccessToken(ctx, granterDID, at.Identifier, at.DocumentIdentifier, relationship.Document.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	if response == nil || response.Document == nil {
		return nil, documents.ErrDocumentInvalid
	}

	model, err := s.Service.DeriveFromCoreDocument(*response.Document)
	if err != nil {
		return nil, err
	}

	if err := s.receivedEntityValidator().Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return model, nil
}

// New returns a new uninitialised Entity.
func (s service) New(_ string) (documents.Document, error) {
	return new(Entity), nil
}

// Validate takes care of entity validation
func (s service) Validate(ctx context.Context, model documents.Document, old documents.Document) error {
	return fieldValidator(s.factory).Validate(old, model)
}
