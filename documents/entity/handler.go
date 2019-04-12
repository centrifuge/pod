package entity

import (
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("entity-api")

// grpcHandler handles all the entity document related actions
// anchoring, sending, finding stored entity document
type grpcHandler struct {
	service Service
	config  config.Service
}

// GRPCHandler returns an implementation of entity.DocumentServiceServer
func GRPCHandler(config config.Service, srv Service) cliententitypb.EntityServiceServer {
	return &grpcHandler{
		service: srv,
		config:  config,
	}
}

// Create handles the creation of the entities and anchoring the documents on chain
func (h *grpcHandler) Create(ctx context.Context, req *cliententitypb.EntityCreatePayload) (*cliententitypb.EntityResponse, error) {
	apiLog.Debugf("Create request %v", req)
	cctx, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	m, err := h.service.DeriveFromCreatePayload(cctx, req)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive create payload")
	}

	// validate and persist
	m, jobID, _, err := h.service.Create(cctx, m)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not create document")
	}

	resp, err := h.service.DeriveEntityResponse(m)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}

// Update handles the document update and anchoring
func (h *grpcHandler) Update(ctx context.Context, payload *cliententitypb.EntityUpdatePayload) (*cliententitypb.EntityResponse, error) {
	apiLog.Debugf("Update request %v", payload)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	doc, err := h.service.DeriveFromUpdatePayload(ctxHeader, payload)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive update payload")
	}

	doc, jobID, _, err := h.service.Update(ctxHeader, doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not update document")
	}

	resp, err := h.service.DeriveEntityResponse(doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}

// GetVersion returns the requested version of the document
func (h *grpcHandler) GetVersion(ctx context.Context, getVersionRequest *cliententitypb.GetVersionRequest) (*cliententitypb.EntityResponse, error) {
	apiLog.Debugf("Get version request %v", getVersionRequest)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(getVersionRequest.Identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "identifier is invalid")
	}

	version, err := hexutil.Decode(getVersionRequest.Version)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "version is invalid")
	}

	model, err := h.service.GetVersion(ctxHeader, identifier, version)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
	}

	resp, err := h.service.DeriveEntityResponse(model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}

// Get returns the entity the latest version of the document with given identifier
func (h *grpcHandler) Get(ctx context.Context, getRequest *cliententitypb.GetRequest) (*cliententitypb.EntityResponse, error) {
	apiLog.Debugf("Get request %v", getRequest)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(getRequest.Identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "identifier is an invalid hex string")
	}

	model, err := h.service.GetCurrentVersion(ctxHeader, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
	}

	resp, err := h.service.DeriveEntityResponse(model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}

// GetEntityByRelationship returns the entity model from database or requests from granter
func (h *grpcHandler) GetEntityByRelationship(ctx context.Context, getRequest *cliententitypb.GetRequestRelationship) (*cliententitypb.EntityResponse, error) {
	apiLog.Debugf("Get request %v", getRequest)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	relationshipIdentifier, err := hexutil.Decode(getRequest.RelationshipIdentifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "identifier is an invalid hex string")
	}

	model, err := h.service.GetEntityByRelationship(ctxHeader, relationshipIdentifier)
	if err != nil {
		return nil, err
	}

	resp, err := h.service.DeriveEntityResponse(model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}

// ListEntityRelationships lists all the relationships associated with the passed in entity identifier
func (h *grpcHandler) ListEntityRelationships(ctx context.Context, getRequest *cliententitypb.GetRequest) (*cliententitypb.RelationshipResponse, error) {
	apiLog.Debugf("Get request %v", getRequest)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	entityIdentifier, err := hexutil.Decode(getRequest.Identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "identifier is an invalid hex string")
	}

	entity, relationships, err := h.service.ListEntityRelationships(ctxHeader, entityIdentifier)
	if err != nil {
		return nil, err
	}

	resp, err := h.service.DeriveRelationshipsListResponse(entity, relationships)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}

func (h *grpcHandler) Share(ctx context.Context, req *cliententitypb.RelationshipPayload) (*cliententitypb.RelationshipResponse, error) {
	apiLog.Debugf("Share request %v", req)
	cctx, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	m, err := h.service.DeriveFromSharePayload(cctx, req)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive share payload")
	}

	// validate and persist
	m, jobID, _, err := h.service.Share(cctx, m)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not create document")
	}

	resp, err := h.service.DeriveEntityRelationshipResponse(m)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}

func (h *grpcHandler) Revoke(ctx context.Context, payload *cliententitypb.RelationshipPayload) (*cliententitypb.RelationshipResponse, error) {
	apiLog.Debugf("Revoke request %v", payload)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	m, err := h.service.DeriveFromRevokePayload(ctxHeader, payload)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive revoke payload")
	}

	m, jobID, _, err := h.service.Revoke(ctxHeader, m)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not update document")
	}

	resp, err := h.service.DeriveEntityRelationshipResponse(m)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}
