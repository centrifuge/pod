package api

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
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
func registerServices(ctx context.Context, cfg Config, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// node object registry
	nodeObjReg, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to get %s", bootstrap.NodeObjRegistry)
	}

	// load dependencies
	registry, ok := nodeObjReg[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return fmt.Errorf("failed to get %s", documents.BootstrappedRegistry)
	}
	payObService, ok := nodeObjReg[nft.BootstrappedPayObService].(nft.PaymentObligation)
	if !ok {
		return fmt.Errorf("failed to get %s", nft.BootstrappedPayObService)
	}

	// documents (common)
	documentpb.RegisterDocumentServiceServer(grpcServer, documents.GRPCHandler(registry))
	err := documentpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// invoice
	invCfg := cfg.(config.Config)
	handler, err := invoice.GRPCHandler(invCfg, registry)
	if err != nil {
		return err
	}
	invoicepb.RegisterDocumentServiceServer(grpcServer, handler)
	err = invoicepb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// purchase orders
	poCfg := cfg.(config.Config)
	srv, err := purchaseorder.GRPCHandler(poCfg, registry)
	if err != nil {
		return fmt.Errorf("failed to get purchase order handler: %v", err)
	}

	purchaseorderpb.RegisterDocumentServiceServer(grpcServer, srv)
	err = purchaseorderpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// healthcheck
	hcCfg := cfg.(healthcheck.Config)
	healthpb.RegisterHealthCheckServiceServer(grpcServer, healthcheck.GRPCHandler(hcCfg))
	err = healthpb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	nftpb.RegisterNFTServiceServer(grpcServer, nft.GRPCHandler(payObService))
	err = nftpb.RegisterNFTServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	return nil
}
