package entityrelationship

import (
	"context"
	"encoding/json"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

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
func (s service) Create(ctx context.Context, relationship documents.Model) (documents.Model, jobs.JobID, chan error, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	relationship, err = s.validateAndPersist(ctx, nil, relationship, CreateValidator(s.factory))
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}

	jobID := contextutil.Job(ctx)
	jobID, done, err := documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, selfDID, jobID, relationship.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return relationship, jobID, done, nil
}

// Update finds the old document, validates the new version and persists the updated document
// For Entity Relationships, Update encompasses the Revoke functionality from the Entity Client API endpoint
func (s service) Update(ctx context.Context, updated documents.Model) (documents.Model, jobs.JobID, chan error, error) {
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
	jobID, done, err := documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, selfDID, jobID, updated.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return updated, jobID, done, nil
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

// CreateModel creates entity relationship from the payload, validates, persists, and returns the document.
func (s service) CreateModel(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	e := new(EntityRelationship)
	if err := e.unpackFromCreatePayload(ctx, payload); err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate invoice
	err := CreateValidator(s.factory).Validate(nil, e)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	did := *e.Data.OwnerIdentity
	err = s.repo.Create(did[:], e.CurrentVersion(), e)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, e.CurrentVersion())
	return e, jobID, err
}

// UpdateModel revokes the entity relationship of a target identity.
func (s service) UpdateModel(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	var data Data
	err := json.Unmarshal(payload.Data, &data)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	id, err := s.repo.FindEntityRelationshipIdentifier(data.EntityIdentifier, *data.OwnerIdentity, *data.TargetIdentity)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	r, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	er := new(EntityRelationship)
	err = er.revokeRelationship(r.(*EntityRelationship), *data.TargetIdentity)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	// validate invoice
	err = UpdateValidator(s.factory, s.anchorRepo).Validate(r, er)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	did := *er.Data.OwnerIdentity
	err = s.repo.Create(did[:], er.CurrentVersion(), er)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, er.CurrentVersion())
	return er, jobID, err
}

// TODO
func (s service) Derive(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error) {
	return nil, errors.New("not implemented")
}

// TODO
func (s service) Patch(ctx context.Context, model documents.Model, payload documents.UpdatePayload) (documents.Model, error) {
	return nil, errors.New("not implemented")
}

// Validate takes care of document validation
func (s service) Validate(ctx context.Context, model documents.Model) error {
	return fieldValidator(s.factory).Validate(nil, model)
}
