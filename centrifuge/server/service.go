package server

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/controller"
)

// RegisterServices registers all endpoints to the grpc server
func RegisterServices(grpcServer *grpc.Server, ctx context.Context, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) {
	invoicepb.RegisterInvoiceDocumentServiceServer(grpcServer, &invoicecontroller.InvoiceDocumentController{})
	err := invoicepb.RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}

}
