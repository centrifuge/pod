package userapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// Service provides functionality for User APIs.
type Service struct {
	coreAPISrv            coreapi.Service
	entityRelationshipSrv entityrelationship.Service
	entitySrv             entity.Service
	config                config.Service
}

func convertEntityRequest(req CreateEntityRequest) (documents.CreatePayload, error) {
	coreAPIReq := coreapi.CreateDocumentRequest{
		Scheme:      entity.Scheme,
		WriteAccess: req.WriteAccess,
		ReadAccess:  req.ReadAccess,
		Data:        req.Data,
		Attributes:  req.Attributes,
	}

	return coreapi.ToDocumentsCreatePayload(coreAPIReq)
}

// CreateEntity creates Entity document and anchors it.
func (s Service) CreateEntity(ctx context.Context, req CreateEntityRequest) (documents.Model, jobs.JobID, error) {
	docReq, err := convertEntityRequest(req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.CreateDocument(ctx, docReq)
}

// UpdateEntity updates existing entity associated with docID  with provided data and anchors it.
func (s Service) UpdateEntity(ctx context.Context, docID []byte, req CreateEntityRequest) (documents.Model, jobs.JobID, error) {
	docReq, err := convertEntityRequest(req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.UpdateDocument(ctx, documents.UpdatePayload{
		DocumentID:    docID,
		CreatePayload: docReq,
	})
}

// GetEntity returns the Entity associated with docID.
func (s Service) GetEntity(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.coreAPISrv.GetDocument(ctx, docID)
}

// ShareEntity shares an entity relationship document with target identity.
func (s Service) ShareEntity(ctx context.Context, docID []byte, req ShareEntityRequest) (documents.Model, jobs.JobID, error) {
	r, err := convertShareEntityRequest(ctx, docID, req.TargetIdentity)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.CreateDocument(ctx, r)
}

// RevokeRelationship revokes target_identity's access to entity.
func (s Service) RevokeRelationship(ctx context.Context, docID []byte, req ShareEntityRequest) (documents.Model, jobs.JobID, error) {
	r, err := convertShareEntityRequest(ctx, docID, req.TargetIdentity)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.UpdateDocument(ctx, documents.UpdatePayload{
		DocumentID:    docID,
		CreatePayload: r,
	})
}

// GetEntityByRelationship returns an entity through a relationship ID.
func (s Service) GetEntityByRelationship(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.entitySrv.GetEntityByRelationship(ctx, docID)
}
