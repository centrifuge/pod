package entityrelationship

import (
	"bytes"
	"context"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for entity
type Service interface {
	documents.Service

	// DeriverFromPayload derives Entity from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *entitypb.RelationshipPayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives entity model from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *entitypb.RelationshipPayload) (documents.Model, error)

	// DeriveEntityRelationshipData returns the entity relationship data as client data
	DeriveEntityRelationshipData(entity documents.Model) (*entitypb.RelationshipData, error)

	// DeriveEntityRelationshipResponse returns the entity relationship model in our standard client format
	DeriveEntityRelationshipResponse(entity documents.Model) (*entitypb.RelationshipResponse, error)

	// GetEntityRelation returns a entity relation based on an entity id
	GetEntityRelationship(entityID, version []byte) (*EntityRelationship, error)
}

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo      repository
	queueSrv  queue.TaskQueuer
	txManager transactions.Manager
	factory   identity.Factory
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo repository,
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
	er := new(EntityRelationship)
	err := er.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return er, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
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
		OwnerIdentity:  owner,
		TargetIdentity: payload.TargetIdentity,
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
	fmt.Println("in repo!!!", s.repo, selfDID[:], er.CurrentVersion(), er)
	err = s.repo.Create(selfDID[:], er.CurrentVersion(), er)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return er, nil
}

// Create takes and entity model and does required validation checks, tries to persist to DB
// For Entity Relationships, Create encompasses the Revoke functionality from the Entity Client API endpoint
func (s service) Create(ctx context.Context, relationship documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	relationship, err = s.validateAndPersist(ctx, nil, relationship, CreateValidator(s.factory))
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}

	txID := contextutil.TX(ctx)
	txID, done, err := documents.CreateAnchorTransaction(s.txManager, s.queueSrv, selfDID, txID, relationship.CurrentVersion())
	if err != nil {
		return nil, transactions.NilTxID(), nil, err
	}
	return relationship, txID, done, nil
}

// Update finds the old document, validates the new version and persists the updated document
// For Entity Relationships, Update encompasses the Revoke functionality from the Entity Client API endpoint
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
		Relationship: data,
	}, nil

}

// DeriveEntityRelationshipData returns create response from entity model
func (s service) DeriveEntityRelationshipData(doc documents.Model) (*entitypb.RelationshipData, error) {
	er, ok := doc.(*EntityRelationship)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return er.getClientData(), nil
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
	fmt.Println(did, eID)
	r, err := s.repo.FindEntityRelationship(eID, *did[0])
	if err != nil {
		return nil, err
	}
	r.CoreDocument, err = r.DeleteAccessToken(ctx, payload.TargetIdentity)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Get returns the entity relationship requested
func (s service) GetEntityRelationship(entityID, version []byte) (*EntityRelationship, error) {
	if entityID == nil {
		return &EntityRelationship{}, documents.ErrPayloadNil
	}

	relationships, err := s.repo.ListAllRelationships(entityID)
	if err != nil {
		return &EntityRelationship{}, err
	}

	if version != nil {
		for _, r := range relationships {
			if bytes.Equal(r.Document.CurrentVersion, version) {
				return &r, nil
			}
		}
	}
	return &EntityRelationship{}, documents.ErrDocumentNotFound
}
