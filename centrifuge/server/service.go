package server

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/controller"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/controller"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// RegisterServices registers all endpoints to the grpc server
func RegisterServices(grpcServer *grpc.Server, ctx context.Context, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) {
	invoicepb.RegisterInvoiceDocumentServiceServer(grpcServer, &invoicecontroller.InvoiceDocumentController{})
	err := invoicepb.RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}
	purchaseorderpb.RegisterPurchaseOrderDocumentServiceServer(grpcServer, &purchaseordercontroller.PurchaseOrderDocumentController{})
	err = purchaseorderpb.RegisterPurchaseOrderDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}

	healthpb.RegisterHealthCheckServiceServer(grpcServer, &healthCheckController{})
	err = healthpb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}
}

// healthCheckController interfaces the grpc health check calls
type healthCheckController struct{}

func (hcc *healthCheckController) Ping(ctx context.Context, empty *empty.Empty) (pong *healthpb.Pong, err error) {
	pong = new(healthpb.Pong)
	pong.Message = fmt.Sprintf("%s+%s", version.GetVersion(), config.Config.GetNetworkString())
	return pong, nil
}
