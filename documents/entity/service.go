package entity

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// DeriveFromPayload derives Entity from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives entity model from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityUpdatePayload) (documents.Model, error)

	// DeriveEntityData returns the entity data as client data
	DeriveEntityData(entity documents.Model) (*cliententitypb.EntityData, error)

	// DeriveEntityResponse returns the entity model in our standard client format
	DeriveEntityResponse(ctx context.Context, entity documents.Model) (*cliententitypb.EntityResponse, error)

	// ListEntityRelationships lists all the relationships associated with the passed in entity identifier
	ListEntityRelationships(ctx context.Context, entityIdentifier []byte) (documents.Model, []documents.Model, error)

	// GetEntityByRelationship returns the entity model from database or requests from granter
	GetEntityByRelationship(ctx context.Context, relationshipIdentifier []byte) (documents.Model, error)

	// DeriveFromSharePayload derives the entity relationship from the relationship payload
	DeriveFromSharePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error)

	// Share takes an entity relationship, validates it, and tries to persist it to the DB
	Share(ctx context.Context, entityRelationship documents.Model) (documents.Model, jobs.JobID, chan bool, error)

	// DeriveFromRevokePayload derives the revoked entity relationship from the relationship payload
	DeriveFromRevokePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error)

	// Revoke takes a revoked entity relationship, validates it, and tries to persist it to the DB
	Revoke(ctx context.Context, entityRelationship documents.Model) (documents.Model, jobs.JobID, chan bool, error)

	// DeriveEntityRelationshipResponse returns create response from entity relationship model
	DeriveEntityRelationshipResponse(model documents.Model) (*cliententitypb.RelationshipResponse, error)
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo                    documents.Repository
	queueSrv                queue.TaskQueuer
	jobManager              jobs.Manager
	factory                 identity.Factory
	processor               documents.DocumentRequestProcessor
	erService               entityrelationship.Service
	anchorRepo              anchors.AnchorRepository
	idService               identity.ServiceDID
	receivedEntityValidator func() documents.ValidatorGroup
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager,
	factory identity.Factory,
	erService entityrelationship.Service,
	idService identity.ServiceDID,
	anchorRepo anchors.AnchorRepository,
	processor documents.DocumentRequestProcessor,
	receivedEntityValidator func() documents.ValidatorGroup,
) Service {
	return service{
		repo:                    repo,
		queueSrv:                queueSrv,
		jobManager:              jobManager,
		Service:                 srv,
		factory:                 factory,
		erService:               erService,
		idService:               idService,
		anchorRepo:              anchorRepo,
		processor:               processor,
		receivedEntityValidator: receivedEntityValidator,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an entity
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Model, error) {
	entity := new(Entity)
	err := entity.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return entity, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrPayloadNil
	}

	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigAccountID
	}

	entity := new(Entity)
	err = entity.InitEntityInput(payload, did)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return entity, nil
}

// validateAndPersist validates the document, calculates the data root, and persists to DB
func (s service) validateAndPersist(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
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

// Create takes an entity model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, entity documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	entity, err = s.validateAndPersist(ctx, nil, entity, CreateValidator(s.factory))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(s.jobManager, s.queueSrv, selfDID, jobID, entity.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return entity, jobID, done, nil
}

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx context.Context, new documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	old, err := s.GetCurrentVersion(ctx, new.ID())
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	new, err = s.validateAndPersist(ctx, old, new, UpdateValidator(s.factory, s.anchorRepo))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(s.jobManager, s.queueSrv, selfDID, jobID, new.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return new, jobID, done, nil
}

// DeriveEntityResponse returns create response from entity model
func (s service) DeriveEntityResponse(ctx context.Context, model documents.Model) (*cliententitypb.EntityResponse, error) {
	data, err := s.DeriveEntityData(model)
	if err != nil {
		return nil, err
	}

	// note that token registry is(must be) irrelevant here
	h, err := documents.DeriveResponseHeader(nil, model)
	if err != nil {
		return nil, err
	}

	entityID := model.ID()
	var relationships []*cliententitypb.Relationship
	// if this identity is not the owner of the entity, return an empty relationships array
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.New("failed to get self ID")
	}

	isCollaborator, err := model.IsDIDCollaborator(selfDID)
	if err != nil {
		return nil, err
	}
	if !isCollaborator {
		return &cliententitypb.EntityResponse{
			Header: h,
			Data: &cliententitypb.EntityDataResponse{
				Entity:        data,
				Relationships: relationships,
			},
		}, nil
	}
	_, models, err := s.ListEntityRelationships(ctx, entityID)
	if err != nil {
		return nil, err
	}

	//list the relationships associated with the entity
	if models != nil {
		for _, m := range models {
			tokens, err := m.GetAccessTokens()
			if err != nil {
				return nil, err
			}

			targetDID := m.(*entityrelationship.EntityRelationship).TargetIdentity.String()
			r := &cliententitypb.Relationship{
				Identity: targetDID,
				Active:   len(tokens) != 0,
			}
			relationships = append(relationships, r)
		}
	}

	return &cliententitypb.EntityResponse{
		Header: h,
		Data: &cliententitypb.EntityDataResponse{
			Entity:        data,
			Relationships: relationships,
		},
	}, nil
}

// DeriveEntityRelationshipData returns the relationship data from an entity relationship model
func (s service) DeriveEntityRelationshipData(model documents.Model) (*cliententitypb.RelationshipData, error) {
	return s.erService.DeriveEntityRelationshipData(model)
}

// DeriveEntityData returns create response from entity model
func (s service) DeriveEntityData(doc documents.Model) (*cliententitypb.EntityData, error) {
	entity, ok := doc.(*Entity)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return entity.getClientData()
}

// DeriveFromUpdatePayload returns a new version of the old entity identified by identifier in payload
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityUpdatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrPayloadNil
	}

	// get latest old version of the document
	id, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentIdentifier, errors.New("failed to decode identifier: %v", err))
	}

	old, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	cs, err := documents.FromClientCollaboratorAccess(payload.ReadAccess, payload.WriteAccess)
	if err != nil {
		return nil, err
	}

	entity := new(Entity)
	err = entity.PrepareNewVersion(old, payload.Data, cs)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPrepareCoreDocument, errors.New("failed to load entity from data: %v", err))
	}

	return entity, nil
}

// GetEntityByRelationship returns the entity model from database or requests from a granter peer
func (s service) GetEntityByRelationship(ctx context.Context, relationshipIdentifier []byte) (documents.Model, error) {
	model, err := s.erService.GetCurrentVersion(ctx, relationshipIdentifier)
	if err != nil {
		return nil, entityrelationship.ErrERNotFound
	}

	relationship, ok := model.(*entityrelationship.EntityRelationship)
	if !ok {
		return nil, entityrelationship.ErrNotEntityRelationship
	}
	// TODO: to be enabled with document syncing
	//entityIdentifier := relationship.EntityIdentifier

	//if s.Service.Exists(ctx, entityIdentifier) {
	//	entity, err := s.Service.GetCurrentVersion(ctx, entityIdentifier)
	//	if err != nil {
	//		// in case of an error try to get document from collaborator
	//		return s.requestEntityWithRelationship(ctx, relationship)
	//	}
	//
	//	// check if stored document is the latest version
	//	if err := documents.LatestVersionValidator(s.anchorRepo).Validate(nil, entity); err != nil {
	//		return s.requestEntityWithRelationship(ctx, relationship)
	//	}
	//
	//	return entity, nil
	//}
	return s.requestEntityWithRelationship(ctx, relationship)
}

// ListEntityRelationships lists all the latest versions of the relationships associated with the passed in entity identifier
func (s service) ListEntityRelationships(ctx context.Context, entityIdentifier []byte) (documents.Model, []documents.Model, error) {
	entity, err := s.GetCurrentVersion(ctx, entityIdentifier)
	if err != nil {
		return nil, nil, err
	}
	relationships, err := s.erService.GetEntityRelationships(ctx, entityIdentifier)
	if err != nil {
		return nil, nil, err
	}
	return entity, relationships, nil
}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	var entity documents.Model
	if s.Service.Exists(ctx, documentID) {
		entity, err = s.Service.GetCurrentVersion(ctx, documentID)

		if err != nil {
			return nil, err
		}

		isCollaborator, err := entity.IsDIDCollaborator(selfDID)
		if err != nil {
			return nil, err
		}
		if !isCollaborator {
			return nil, documents.ErrNoCollaborator
		}
		return entity, nil
	}

	return nil, documents.ErrDocumentNotFound
}

func (s service) requestEntityWithRelationship(ctx context.Context, relationship *entityrelationship.EntityRelationship) (documents.Model, error) {
	accessTokens, err := relationship.GetAccessTokens()
	if err != nil {
		return nil, documents.ErrCDAttribute
	}

	// only one access token per entity relationship
	if len(accessTokens) != 1 {
		return nil, entityrelationship.ErrERNoToken
	}

	at := accessTokens[0]
	if !utils.IsSameByteSlice(at.DocumentIdentifier, relationship.EntityIdentifier) {
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

	// TODO: to be enabled with document syncing
	//if err = s.store(ctx, model); err != nil {
	//	return nil, err
	//}

	return model, nil
}

func (s service) store(ctx context.Context, model documents.Model) error {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	if s.Service.Exists(ctx, model.CurrentVersion()) {
		err = s.repo.Update(selfDID[:], model.CurrentVersion(), model)
		if err != nil {
			return errors.NewTypedError(documents.ErrDocumentPersistence, err)
		}
	} else {
		err = s.repo.Create(selfDID[:], model.CurrentVersion(), model)
		if err != nil {
			return errors.NewTypedError(documents.ErrCDCreate, err)
		}
	}
	return nil
}

// DeriveFromSharePayload derives the entity relationship from the relationship payload
func (s service) DeriveFromSharePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	return s.erService.DeriveFromCreatePayload(ctx, payload)
}

// Share takes an entity relationship, validates it, and tries to persist it to the DB
func (s service) Share(ctx context.Context, entityRelationship documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	return s.erService.Create(ctx, entityRelationship)
}

// DeriveFromRevokePayload derives the revoked entity relationship from the relationship payload
func (s service) DeriveFromRevokePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	return s.erService.DeriveFromUpdatePayload(ctx, payload)
}

// Revoke takes a revoked entity relationship, validates it, and tries to persist it to the DB
func (s service) Revoke(ctx context.Context, entityRelationship documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	return s.erService.Update(ctx, entityRelationship)
}

// DeriveEntityRelationshipResponse returns create response from entity relationship model
func (s service) DeriveEntityRelationshipResponse(model documents.Model) (*cliententitypb.RelationshipResponse, error) {
	return s.erService.DeriveEntityRelationshipResponse(model)
}
