package entityrelationship

import (
	"context"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// DeriverFromPayload derives Entity from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityRelationshipCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives entity model from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityRelationshipUpdatePayload) (documents.Model, error)

	// DeriveEntityRelationshipData returns the entity relationship data as client data
	DeriveEntityRelationshipData(entity documents.Model) (*cliententitypb.EntityRelationshipData, error)

	// DeriveEntityRelationshipResponse returns the entity relationship model in our standard client format
	DeriveEntityRelationshipResponse(entity documents.Model) (*cliententitypb.EntityRelationshipResponse, error)
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	storage   storage.Repository
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
	storage storage.Repository,
) Service {
	return service{
		repo:      repo,
		queueSrv:  queueSrv,
		txManager: txManager,
		Service:   srv,
		factory:   factory,
		storage:   storage,
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

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityRelationshipCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	er := new(EntityRelationship)
	err := er.InitEntityRelationshipInput(payload)
	if err != nil {
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

// Create takes and entity model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, entityRelationship documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	entityRelationship, err = s.validateAndPersist(ctx, nil, entityRelationship, CreateValidator(s.factory))
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, selfDID, txID, entityRelationship.CurrentVersion())
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}
	return entityRelationship, txID, done, nil
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
func (s service) DeriveEntityRelationshipResponse(model documents.Model) (*cliententitypb.EntityRelationshipResponse, error) {
	data, err := s.DeriveEntityRelationshipData(model)
	if err != nil {
		return nil, err
	}

	cs, err := model.GetCollaborators()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrFailedCollaborators, err)
	}

	var css []string
	for _, c := range cs {
		css = append(css, c.String())
	}

	h := &cliententitypb.ResponseHeader{
		DocumentId:    hexutil.Encode(model.ID()),
		VersionId:     hexutil.Encode(model.CurrentVersion()),
		Collaborators: css,
	}

	return &cliententitypb.EntityRelationshipResponse{
		Header: h,
		Data:   data,
	}, nil

}

// DeriveEntityData returns create response from entity model
func (s service) DeriveEntityRelationshipData(doc documents.Model) (*cliententitypb.EntityRelationshipData, error) {
	er, ok := doc.(*EntityRelationship)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return er.getClientData(), nil
}

// DeriveFromUpdatePayload returns a new version of the old entity identified by identifier in payload
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityRelationshipUpdatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}
	dids, err := identity.StringsToDIDs(payload.Data.OwnerIdentity, payload.Data.TargetIdentity)
	if err != nil {
		return nil, err
	}
	models, err := s.storage.GetAllByPrefix(payload.Data.OwnerIdentity)
	if err != nil {
		return nil, err
	}
	var old documents.Model
	for _, m := range models {
		if m.(*EntityRelationship).TargetIdentity == dids[1] {
			old = m.(*EntityRelationship)
		}
	}
	er := new(EntityRelationship)
	err = er.PrepareNewVersion(old, payload.Data, nil)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPrepareCoreDocument, errors.New("failed to load entity from data: %v", err))
	}

	return er, nil
}
