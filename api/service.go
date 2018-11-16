package api

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/healthcheck"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/documents"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/health"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// registerServices registers all endpoints to the grpc server
func registerServices(ctx context.Context, config Config, registry *documents.ServiceRegistry, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// documents (common)
	documentpb.RegisterDocumentServiceServer(grpcServer, documents.GRPCHandler(registry))
	err := documentpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// invoice
	handler, err := invoice.GRPCHandler(registry)
	if err != nil {
		return err
	}
	invoicepb.RegisterDocumentServiceServer(grpcServer, handler)
	err = invoicepb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// purchase orders
	srv, err := purchaseorder.GRPCHandler(registry)
	if err != nil {
		return fmt.Errorf("failed to get purchase order handler: %v", err)
	}

	purchaseorderpb.RegisterDocumentServiceServer(grpcServer, srv)
	err = purchaseorderpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// healthcheck
	healthpb.RegisterHealthCheckServiceServer(grpcServer, healthcheck.GRPCHandler(config))
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
