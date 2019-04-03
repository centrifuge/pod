package entity

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
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
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo      documents.Repository
	queueSrv  queue.TaskQueuer
	txManager transactions.Manager
	factory   identity.Factory
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	queueSrv queue.TaskQueuer,
	txManager transactions.Manager,
	factory identity.Factory,
) Service {
	return service{
		repo:      repo,
		queueSrv:  queueSrv,
		txManager: txManager,
		Service:   srv,
		factory:   factory,
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
		return nil, documents.ErrDocumentNil
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

// Create takes and entity model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, entity documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	entity, err = s.validateAndPersist(ctx, nil, entity, CreateValidator(s.factory))
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, selfDID, txID, entity.CurrentVersion())
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}
	return entity, txID, done, nil
}

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx context.Context, new documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	old, err := s.GetCurrentVersion(ctx, new.ID())
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	new, err = s.validateAndPersist(ctx, old, new, UpdateValidator(s.factory))
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, selfDID, txID, new.CurrentVersion())
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}
	return new, txID, done, nil
}

// DeriveEntityResponse returns create response from entity model
func (s service) DeriveEntityResponse(model documents.Model) (*cliententitypb.EntityResponse, error) {
	data, err := s.DeriveEntityData(model)
	if err != nil {
		return nil, err
	}

	h, err := documents.DeriveResponseHeader(model)
	if err != nil {
		return nil, err
	}

	return &cliententitypb.EntityResponse{
		Header: h,
		Data:   data,
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
		return nil, documents.ErrDocumentNil
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

// isDIDaCollaborator returns true if the did is a collaborator of the document
func (s service) isDIDaCollaborator(ctx context.Context, did identity.DID, model documents.Model) (bool, error) {
	collAccess, err := model.GetCollaborators()
	if err != nil {
		return false, err
	}

	for _, d := range collAccess.ReadWriteCollaborators {
		if d == did {
			return true, nil
		}
	}
	for _, d := range collAccess.ReadCollaborators {
		if d == did {
			return true, nil
		}
	}
	return false, nil
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
		isCollaborator, err = s.isDIDaCollaborator(ctx, selfDID, entity)
		if err != nil {
			return nil, err
		}
	}
	if isCollaborator {
		// todo add relationship array
		return entity, nil
	}

	// todo call entityRelationship service and request Entity document from other collaborators
	return nil, documents.ErrDocumentNotFound
}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	return s.get(ctx, documentID, nil)

}
