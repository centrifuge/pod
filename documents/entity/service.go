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
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// GetEntityByRelationship returns the entity model from database or requests from granter
	GetEntityByRelationship(ctx context.Context, relationshipIdentifier []byte) (documents.Model, error)
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
	jobID, done, err := documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, selfDID, jobID, new.CurrentVersion())
	if err != nil {
		return nil, jobs.NilJobID(), nil, err
	}
	return new, jobID, done, nil
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

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
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

// CreateModel creates entity from the payload, validates, persists, and returns the entity.
func (s service) CreateModel(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	if payload.Data == nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentNil
	}

	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	e := new(Entity)
	if err := e.unpackFromCreatePayload(did, payload); err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate invoice
	err = CreateValidator(s.factory).Validate(nil, e)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(did[:], e.CurrentVersion(), e)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, e.CurrentVersion())
	return e, jobID, err
}

// UpdateModel updates the migrates the current entity to next version with data from the update payload
func (s service) UpdateModel(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	if payload.Data == nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentNil
	}

	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentConfigAccountID
	}

	old, err := s.GetCurrentVersion(ctx, payload.DocumentID)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	oldEntity, ok := old.(*Entity)
	if !ok {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("%v is not an Entity", hexutil.Encode(payload.DocumentID)))
	}

	e := new(Entity)
	err = e.unpackFromUpdatePayload(oldEntity, payload)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	err = UpdateValidator(s.factory, s.anchorRepo).Validate(old, e)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	err = s.repo.Create(did[:], e.CurrentVersion(), e)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = documents.CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, e.CurrentVersion())
	return e, jobID, err
}
