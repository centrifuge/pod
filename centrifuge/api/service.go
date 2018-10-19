package api

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/healthcheck/controller"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	legacyInvoice "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/invoice"
	legacyPurchaseOrder "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
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
	handler, err := invoice.GRPCHandler()
	if err != nil {
		return err
	}
	invoicepb.RegisterDocumentServiceServer(grpcServer, handler)
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
	purchaseorderpb.RegisterDocumentServiceServer(grpcServer, purchaseorder.GRPCHandler())
	err = purchaseorderpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}
	legacyPurchaseOrder.RegisterPurchaseOrderDocumentServiceServer(grpcServer, purchaseorder.LegacyGRPCHandler(
		purchaseorder.GetLegacyRepository(),
		coredocumentprocessor.DefaultProcessor(identity.IDService, p2p.NewP2PClient(), anchors.GetAnchorRepository())))
	err = legacyPurchaseOrder.RegisterPurchaseOrderDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// healthcheck
	healthpb.RegisterHealthCheckServiceServer(grpcServer, &healthcheckcontroller.HealthCheckController{})
	err = healthpb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	nftpb.RegisterNFTServiceServer(grpcServer, nft.GRPCHandler())
	err = nftpb.RegisterNFTServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	return nil
}
