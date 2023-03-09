package entity

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/documents/entityrelationship"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/pallets/anchors"
	"github.com/centrifuge/pod/utils"
)

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

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
	identityService         v2.Service
	processor               documents.AnchorProcessor
	erService               entityrelationship.Service
	anchorSrv               anchors.API
	receivedEntityValidator func() documents.Validator
}

func NewService(
	srv documents.Service,
	repo documents.Repository,
	identityService v2.Service,
	erService entityrelationship.Service,
	anchorSrv anchors.API,
	processor documents.AnchorProcessor,
	receivedEntityValidator func() documents.Validator,
) Service {
	return service{
		repo:                    repo,
		Service:                 srv,
		identityService:         identityService,
		erService:               erService,
		anchorSrv:               anchorSrv,
		processor:               processor,
		receivedEntityValidator: receivedEntityValidator,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an entity
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Document, error) {
	entity := new(Entity)

	err := entity.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
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
	identity, err := contextutil.Identity(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrAccountNotFoundInContext, err)
	}

	entity, err := s.Service.GetCurrentVersion(ctx, documentID)
	if err != nil {
		return nil, documents.ErrDocumentNotFound
	}

	isCollaborator, err := entity.IsCollaborator(identity)
	if err != nil || !isCollaborator {
		return nil, ErrIdentityNotACollaborator
	}

	return entity, nil
}

func (s service) requestEntityWithRelationship(ctx context.Context, relationship *entityrelationship.EntityRelationship) (documents.Document, error) {
	accessTokens := relationship.GetAccessTokens()

	// only one access token per entity relationship
	if len(accessTokens) != 1 {
		return nil, entityrelationship.ErrERNoToken
	}

	at := accessTokens[0]
	if !utils.IsSameByteSlice(at.DocumentIdentifier, relationship.Data.EntityIdentifier) {
		return nil, entityrelationship.ErrERInvalidIdentifier
	}

	granterAccountID, err := types.NewAccountID(at.Granter)
	if err != nil {
		return nil, documents.ErrGranterInvalidAccountID
	}

	response, err := s.processor.RequestDocumentWithAccessToken(
		ctx,
		granterAccountID,
		at.Identifier,
		at.DocumentIdentifier,
		relationship.Document.DocumentIdentifier,
	)

	if err != nil {
		return nil, errors.NewTypedError(ErrP2PDocumentRequest, err)
	}

	if response == nil || response.Document == nil {
		return nil, documents.ErrDocumentInvalid
	}

	model, err := s.Service.DeriveFromCoreDocument(response.Document)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentDerive, err)
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
func (s service) Validate(_ context.Context, model documents.Document, old documents.Document) error {
	return fieldValidator(s.identityService).Validate(old, model)
}
