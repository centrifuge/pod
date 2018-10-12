package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("purchaseorder-api")

// grpcHandler handles all the purchase order document related actions
// anchoring, sending, finding stored purchase order document
type grpcHandler struct {
}

// GRPCHandler returns an implementation of clientpurchaseorderpb.PurchaseOrderDocumentServiceServer
func GRPCHandler() clientpurchaseorderpb.DocumentServiceServer {
	return &grpcHandler{}
}

func (grpcHandler) Create(context.Context, *clientpurchaseorderpb.PurchaseOrderCreatePayload) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}

func (grpcHandler) Update(context.Context, *clientpurchaseorderpb.PurchaseOrderUpdatePayload) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}

func (grpcHandler) GetVersion(context.Context, *clientpurchaseorderpb.GetVersionRequest) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}

func (grpcHandler) Get(context.Context, *clientpurchaseorderpb.GetRequest) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}
