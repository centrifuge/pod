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
func GRPCHandler(config config.Service, srv Service) cliententitypb.DocumentServiceServer {
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

	doc, err := h.service.DeriveFromCreatePayload(cctx, req)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive create payload")
	}

	// validate and persist
	doc, txID, _, err := h.service.Create(cctx, doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not create document")
	}

	resp, err := h.service.DeriveEntityResponse(doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.TransactionId = txID.String()
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

	doc, txID, _, err := h.service.Update(ctxHeader, doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not update document")
	}

	resp, err := h.service.DeriveEntityResponse(doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.TransactionId = txID.String()
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
