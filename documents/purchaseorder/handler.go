package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("purchaseorder-api")

// grpcHandler handles all the purchase order document related actions
// anchoring, sending, finding stored purchase order document
type grpcHandler struct {
	service Service
	config  config.Service
}

// GRPCHandler returns an implementation of the purchaseorder DocumentServiceServer
func GRPCHandler(config config.Service, srv Service) clientpurchaseorderpb.PurchaseOrderServiceServer {
	return grpcHandler{
		service: srv,
		config:  config,
	}
}

// Update handles the document update and anchoring
func (h grpcHandler) Update(ctx context.Context, payload *clientpurchaseorderpb.PurchaseOrderUpdatePayload) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
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

	resp, err := h.service.DerivePurchaseOrderResponse(doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.JobId = jobID.String()
	return resp, nil
}
