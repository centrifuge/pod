package invoice

import (
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("invoice-api")

// grpcHandler handles all the invoice document related actions
// anchoring, sending, finding stored invoice document
type grpcHandler struct {
	service Service
	config  configstore.Service
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(config configstore.Service, registry *documents.ServiceRegistry) (clientinvoicepb.DocumentServiceServer, error) {
	srv, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
	if err != nil {
		return nil, err
	}

	return &grpcHandler{
		service: srv.(Service),
		config:  config,
	}, nil
}

// Create handles the creation of the invoices and anchoring the documents on chain
func (h *grpcHandler) Create(ctx context.Context, req *clientinvoicepb.InvoiceCreatePayload) (*clientinvoicepb.InvoiceResponse, error) {
	apiLog.Debugf("Create request %v", req)
	cctx, err := contextutil.CentContext(ctx, h.config)
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
	doc, txID, err := h.service.Create(cctx, doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not create document")
	}

	resp, err := h.service.DeriveInvoiceResponse(doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.TransactionId = txID.String()
	return resp, nil
}

// Update handles the document update and anchoring
func (h *grpcHandler) Update(ctx context.Context, payload *clientinvoicepb.InvoiceUpdatePayload) (*clientinvoicepb.InvoiceResponse, error) {
	apiLog.Debugf("Update request %v", payload)
	ctxHeader, err := contextutil.CentContext(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	doc, err := h.service.DeriveFromUpdatePayload(ctxHeader, payload)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive update payload")
	}

	doc, txID, err := h.service.Update(ctxHeader, doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not update document")
	}

	resp, err := h.service.DeriveInvoiceResponse(doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.TransactionId = txID.String()
	return resp, nil
}

// GetVersion returns the requested version of the document
func (h *grpcHandler) GetVersion(ctx context.Context, getVersionRequest *clientinvoicepb.GetVersionRequest) (*clientinvoicepb.InvoiceResponse, error) {
	apiLog.Debugf("Get version request %v", getVersionRequest)
	ctxHeader, err := contextutil.CentContext(ctx, h.config)
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

	resp, err := h.service.DeriveInvoiceResponse(model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}

// Get returns the invoice the latest version of the document with given identifier
func (h *grpcHandler) Get(ctx context.Context, getRequest *clientinvoicepb.GetRequest) (*clientinvoicepb.InvoiceResponse, error) {
	apiLog.Debugf("Get request %v", getRequest)
	ctxHeader, err := contextutil.CentContext(ctx, h.config)
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

	resp, err := h.service.DeriveInvoiceResponse(model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}
