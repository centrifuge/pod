package funding

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("funding-api")

// grpcHandler handles all the funding extension related actions
type grpcHandler struct {
	service Service
	config  config.Service
}

// GRPCHandler returns an implementation of entity.DocumentServiceServer
func GRPCHandler(config config.Service, srv Service) clientfunpb.FundingServiceServer {
	return &grpcHandler{
		service: srv,
		config:  config,
	}
}

// Get returns a funding agreement from an existing document
func (h *grpcHandler) Get(ctx context.Context, req *clientfunpb.Request) (*clientfunpb.FundingResponse, error) {
	apiLog.Debugf("Get request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.DocumentId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentIdentifier
	}

	model, err := h.service.GetCurrentVersion(ctxHeader, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	resp, err := h.service.DeriveFundingResponse(ctxHeader, model, req.AgreementId)
	if err != nil {
		apiLog.Error(err)
		return nil, extensions.ErrDeriveAttr
	}
	return resp, nil
}

// Sign adds a funding signature to a document
func (h *grpcHandler) Sign(ctx context.Context, req *clientfunpb.Request) (*clientfunpb.FundingResponse, error) {
	apiLog.Debugf("create funding request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.DocumentId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentIdentifier
	}

	// returns model with a signature
	model, err := h.service.Sign(ctxHeader, req.AgreementId, identifier)
	if err != nil {
		return nil, errors.NewTypedError(extensions.ErrPayload, err)
	}

	model, jobID, _, err := h.service.Update(ctxHeader, model)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	resp, err := h.service.DeriveFundingResponse(ctxHeader, model, req.AgreementId)
	if err != nil {
		apiLog.Error(err)
		return nil, errors.NewTypedError(extensions.ErrDeriveAttr, err)
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}

// Get returns a funding agreement from an existing document
func (h *grpcHandler) GetVersion(ctx context.Context, req *clientfunpb.GetVersionRequest) (*clientfunpb.FundingResponse, error) {
	apiLog.Debugf("Get request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.DocumentId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentIdentifier
	}

	version, err := hexutil.Decode(req.VersionId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentVersion
	}

	model, err := h.service.GetVersion(ctxHeader, identifier, version)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentVersionNotFound
	}

	resp, err := h.service.DeriveFundingResponse(ctxHeader, model, req.AgreementId)
	if err != nil {
		apiLog.Error(err)
		return nil, extensions.ErrDeriveAttr
	}
	return resp, nil
}

// GetList returns all funding agreements of a existing document
func (h *grpcHandler) GetList(ctx context.Context, req *clientfunpb.GetListRequest) (*clientfunpb.FundingListResponse, error) {
	apiLog.Debugf("Get request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.DocumentId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentIdentifier
	}

	model, err := h.service.GetCurrentVersion(ctxHeader, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	resp, err := h.service.DeriveFundingListResponse(ctxHeader, model)
	if err != nil {
		apiLog.Error(err)
		return nil, extensions.ErrDeriveAttr
	}
	return resp, nil
}

// GetList returns all funding agreements of a existing document
func (h *grpcHandler) GetListVersion(ctx context.Context, req *clientfunpb.GetListVersionRequest) (*clientfunpb.FundingListResponse, error) {
	apiLog.Debugf("Get request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.DocumentId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentIdentifier
	}

	version, err := hexutil.Decode(req.VersionId)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentVersion
	}

	model, err := h.service.GetVersion(ctxHeader, identifier, version)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	resp, err := h.service.DeriveFundingListResponse(ctxHeader, model)
	if err != nil {
		apiLog.Error(err)
		return nil, extensions.ErrDeriveAttr
	}
	return resp, nil
}

// Update handles an update over an existing funding document extension
func (h *grpcHandler) Update(ctx context.Context, req *clientfunpb.FundingUpdatePayload) (*clientfunpb.FundingResponse, error) {
	apiLog.Debugf("create funding request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	// returns model with updated funding custom fields
	model, err := h.service.DeriveFromUpdatePayload(ctxHeader, req)
	if err != nil {
		return nil, errors.NewTypedError(extensions.ErrPayload, err)
	}

	model, jobID, _, err := h.service.Update(ctxHeader, model)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	resp, err := h.service.DeriveFundingResponse(ctxHeader, model, req.Data.AgreementId)
	if err != nil {
		apiLog.Error(err)
		return nil, errors.NewTypedError(extensions.ErrPayload, err)
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}
