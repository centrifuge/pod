package server

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/grpc"
)

// RegisterServices registers all endpoints to the grpc server
func RegisterServices(grpcServer *grpc.Server, ctx context.Context, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) {
	invoicegrpc.RegisterInvoiceDocumentServiceServer(grpcServer, &invoiceservice.InvoiceDocumentService{})
	err := invoicegrpc.RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}

}
