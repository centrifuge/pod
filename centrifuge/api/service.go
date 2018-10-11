package api

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/healthcheck/controller"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	legacyInvoice "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/purchaseorder/controller"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// registerServices registers all endpoints to the grpc server
func registerServices(ctx context.Context, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// documents (common)
	documentpb.RegisterDocumentServiceServer(grpcServer, documents.GRPCHandler())
	err := documentpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// invoice
	// get the invoice service from the registry, it has to be registered already
	invoiceService, err := documents.GetRegistryInstance().LocateService(documenttypes.InvoiceDataTypeUrl)
	invoicepb.RegisterDocumentServiceServer(grpcServer, invoice.GRPCHandler(invoiceService.(invoice.Service)))
	err = invoicepb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}
	legacyInvoice.RegisterInvoiceDocumentServiceServer(grpcServer, invoice.LegacyGRPCHandler())
	err = legacyInvoice.RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// purchase orders
	purchaseorderpb.RegisterPurchaseOrderDocumentServiceServer(grpcServer, &purchaseordercontroller.PurchaseOrderDocumentController{})
	err = purchaseorderpb.RegisterPurchaseOrderDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// healthcheck
	healthpb.RegisterHealthCheckServiceServer(grpcServer, &healthcheckcontroller.HealthCheckController{})
	err = healthpb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}
	return nil
}
