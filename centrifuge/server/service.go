package server

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/healthcheck"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/handler"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/service"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// RegisterServices registers all endpoints to the grpc server
func RegisterServices(grpcServer *grpc.Server, ctx context.Context, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) {
	invoicepb.RegisterInvoiceDocumentServiceServer(grpcServer, &invoicehandler.Handler{})
	err := invoicepb.RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}
	purchaseorderpb.RegisterPurchaseOrderDocumentServiceServer(grpcServer, &purchaseorderhandler.PurchaseOrderDocumentService{})
	err = purchaseorderpb.RegisterPurchaseOrderDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}

	healthpb.RegisterHealthCheckServiceServer(grpcServer, &healthcheck.Handler{})
	err = healthpb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}
}
