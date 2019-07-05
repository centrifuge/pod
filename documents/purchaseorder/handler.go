package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

// GetVersion returns the requested version of a purchase order
func (h grpcHandler) GetVersion(ctx context.Context, req *clientpurchaseorderpb.GetVersionRequest) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Debugf("GetVersion request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.DocumentId)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "identifier is invalid")
	}

	version, err := hexutil.Decode(req.VersionId)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "version is invalid")
	}

	model, err := h.service.GetVersion(ctxHeader, identifier, version)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
	}

	resp, err := h.service.DerivePurchaseOrderResponse(model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	return resp, nil
}
