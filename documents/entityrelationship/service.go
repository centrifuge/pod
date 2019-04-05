package entityrelationship

import (
	"bytes"
	"context"

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
		return nil, documents.ErrDocumentNil
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

// validateAndPersist is not implemented for the entity relationship
func (s service) validateAndPersist(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	return nil, documents.ErrNotImplemented
}

// Create is not implemented for the EntityRelationship
func (s service) Create(ctx context.Context, entityRelationship documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	return nil, transactions.TxID{}, nil, documents.ErrNotImplemented
}

// Update is not implemented for the EntityRelationship
func (s service) Update(ctx context.Context, new documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	return nil, transactions.TxID{}, nil, documents.ErrNotImplemented
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
		return nil, documents.ErrDocumentNil
	}

	eID, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, err
	}

	did, err := identity.StringsToDIDs(payload.TargetIdentity)
	if err != nil {
		return nil, err
	}

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
