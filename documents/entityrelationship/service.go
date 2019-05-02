package entityrelationship

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// DeriveFromCreatePayload derives Entity Relationship from RelationshipPayload
	DeriveFromCreatePayload(ctx context.Context, payload *entitypb.RelationshipPayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives a revoked entity relationship model from RelationshipPayload
	DeriveFromUpdatePayload(ctx context.Context, payload *entitypb.RelationshipPayload) (documents.Model, error)

	// DeriveEntityRelationshipData returns the entity relationship data as client data
	DeriveEntityRelationshipData(relationship documents.Model) (*entitypb.RelationshipData, error)

	// DeriveEntityRelationshipResponse returns the entity relationship model in our standard client format
	DeriveEntityRelationshipResponse(relationship documents.Model) (*entitypb.RelationshipResponse, error)

	// GetEntityRelationships returns a list of the latest versions of the relevant entity relationship based on an entity id
	GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Model, error)
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo       repository
	queueSrv   queue.TaskQueuer
	jobManager jobs.Manager
	factory    identity.Factory
	anchorRepo anchors.AnchorRepository
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo repository,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager,
	factory identity.Factory,
	anchorRepo anchors.AnchorRepository,
) Service {
	return service{
		repo:       repo,
		queueSrv:   queueSrv,
		jobManager: jobManager,
		Service:    srv,
		factory:    factory,
		anchorRepo: anchorRepo,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an entity
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Model, error) {
	er := new(EntityRelationship)
	err := er.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return er, nil
}

// DeriveFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *entitypb.RelationshipPayload) (documents.Model, error) {
	if payload == nil {
		return nil, documents.ErrPayloadNil
	}

	er := new(EntityRelationship)
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, err
	}
	owner := selfDID.String()
	rd := &entitypb.RelationshipData{
		OwnerIdentity:    owner,
		TargetIdentity:   payload.TargetIdentity,
		EntityIdentifier: payload.Identifier,
	}
	if err = er.InitEntityRelationshipInput(ctx, payload.Identifier, rd); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return er, nil
}

// validateAndPersist validates the document, calculates the data root, and persists to DB
func (s service) validateAndPersist(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	er, ok := new.(*EntityRelationship)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// validate the entity
	err = validator.Validate(old, er)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(selfDID[:], er.CurrentVersion(), er)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return er, nil
}

// Create takes an entity relationship model and does required validation checks, tries to persist to DB
// For Entity Relationships, Create encompasses the Share functionality from the Entity Client API endpoint
func (s service) Create(ctx context.Context, relationship documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	relationship, err = s.validateAndPersist(ctx, nil, relationship, CreateValidator(s.factory))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(s.jobManager, s.queueSrv, selfDID, jobID, relationship.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return relationship, jobID, done, nil
}

// Update finds the old document, validates the new version and persists the updated document
// For Entity Relationships, Update encompasses the Revoke functionality from the Entity Client API endpoint
func (s service) Update(ctx context.Context, updated documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	old, err := s.GetCurrentVersion(ctx, updated.ID())
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	updated, err = s.validateAndPersist(ctx, old, updated, UpdateValidator(s.factory, s.anchorRepo))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(s.jobManager, s.queueSrv, selfDID, jobID, updated.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return updated, jobID, done, nil
}

// DeriveEntityRelationshipResponse returns create response from entity relationship model
func (s service) DeriveEntityRelationshipResponse(model documents.Model) (*entitypb.RelationshipResponse, error) {
	data, err := s.DeriveEntityRelationshipData(model)
	if err != nil {
		return nil, err
	}

	h := &documentpb.ResponseHeader{
		DocumentId: hexutil.Encode(model.ID()),
		Version:    hexutil.Encode(model.CurrentVersion()),
	}

	return &entitypb.RelationshipResponse{
		Header:       h,
		Relationship: []*entitypb.RelationshipData{data},
	}, nil

}

// DeriveEntityRelationshipData returns the relationship data from an entity relationship model
func (s service) DeriveEntityRelationshipData(model documents.Model) (*entitypb.RelationshipData, error) {
	er, ok := model.(*EntityRelationship)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return er.getRelationshipData(), nil
}

// DeriveFromUpdatePayload returns a new version of the indicated Entity Relationship with a deleted access token
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *entitypb.RelationshipPayload) (documents.Model, error) {
	if payload == nil {
		return nil, documents.ErrPayloadNil
	}

	eID, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, err
	}

	did, err := identity.StringsToDIDs(payload.TargetIdentity)
	if err != nil {
		return nil, err
	}

	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.New("failed to get self ID")
	}

	id, err := s.repo.FindEntityRelationshipIdentifier(eID, selfDID, *did[0])
	if err != nil {
		return nil, err
	}

	r, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	cd, err := r.(*EntityRelationship).DeleteAccessToken(ctx, hexutil.Encode(did[0][:]))
	if err != nil {
		return nil, err
	}

	r.(*EntityRelationship).CoreDocument = cd
	return r, nil
}

// GetEntityRelationships returns the latest versions of the entity relationships that involve the entityID passed in
func (s service) GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Model, error) {
	var relationships []documents.Model
	if entityID == nil {
		return nil, documents.ErrPayloadNil
	}

	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.New("failed to get self ID")
	}

	relevant, err := s.repo.ListAllRelationships(entityID, selfDID)
	if err != nil {
		return nil, err
	}

	for _, v := range relevant {
		r, err := s.GetCurrentVersion(ctx, v)
		if err != nil {
			return nil, err
		}
		relationships = append(relationships, r)
	}

	if relationships == nil {
		return nil, nil
	}

	return relationships, nil
}
