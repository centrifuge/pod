package entity

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// DeriverFromPayload derives Entity from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives entity model from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityUpdatePayload) (documents.Model, error)

	// DeriveEntityData returns the entity data as client data
	DeriveEntityData(entity documents.Model) (*cliententitypb.EntityData, error)

	// DeriveEntityResponse returns the entity model in our standard client format
	DeriveEntityResponse(entity documents.Model) (*cliententitypb.EntityResponse, error)

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
	repo      documents.Repository
	queueSrv  queue.TaskQueuer
	txManager jobs.Manager
	factory   identity.Factory
	processor documents.DocumentRequestProcessor
	erService entityrelationship.Service
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	txManager jobs.Manager,
	factory identity.Factory,
	erService entityrelationship.Service,
	processor documents.DocumentRequestProcessor,
) Service {
	return service{
		repo:      repo,
		queueSrv:  queueSrv,
		txManager: txManager,
		Service:   srv,
		factory:   factory,
		erService: erService,
		processor: processor,
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

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, selfDID, txID, entity.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return entity, txID, done, nil
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

	new, err = s.validateAndPersist(ctx, old, new, UpdateValidator(s.factory))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, selfDID, txID, new.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return new, txID, done, nil
}

// DeriveEntityResponse returns create response from entity model
func (s service) DeriveEntityResponse(model documents.Model) (*cliententitypb.EntityResponse, error) {
	data, err := s.DeriveEntityData(model)
	if err != nil {
		return nil, err
	}

	// note that token registry is(must be) irrelevant here
	h, err := documents.DeriveResponseHeader(nil, model)
	if err != nil {
		return nil, err
	}

	return &cliententitypb.EntityResponse{
		Header: h,
		Data: &cliententitypb.EntityDataResponse{
			Entity:        data,
			Relationships: nil,
		},
	}, nil

}

// DeriveEntityData returns create response from entity model
func (s service) DeriveEntityData(doc documents.Model) (*cliententitypb.EntityData, error) {
	entity, ok := doc.(*Entity)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return entity.getClientData(), nil
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

func (s service) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	return s.get(ctx, documentID, version)
}

func (s service) get(ctx context.Context, documentID, version []byte) (documents.Model, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	isCollaborator := false
	var entity documents.Model

	if s.Service.Exists(ctx, documentID) {
		if version == nil {
			entity, err = s.Service.GetCurrentVersion(ctx, documentID)
		} else {
			entity, err = s.Service.GetVersion(ctx, documentID, version)
		}

		if err != nil {
			return nil, err
		}

		isCollaborator, err = entity.IsDIDCollaborator(selfDID)
		if err != nil {
			return nil, err
		}
		if !isCollaborator {
			return nil, documents.ErrNoCollaborator
		}
		// todo add relationship array
		return entity, nil
	}

	return nil, documents.ErrDocumentNotFound
}

func (s service) requestEntityFromCollaborator(documentID, version []byte) (documents.Model, error) {
	/*

		todo steps
		1. Find ER related to Entity document.Identifier
		2. Request document with token s.processor.RequestDocumentWithAccessToken(...) from the first Collaborator
		3. call a new method in documents.Service to validate received document
		4. return entity document if validation
	*/

	return nil, documents.ErrDocumentNotFound
}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	return s.get(ctx, documentID, nil)

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
