package api

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/healthcheck"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/health"
	invoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	purchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// registerServices registers all endpoints to the grpc server
func registerServices(ctx context.Context, cfg Config, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// node object registry
	nodeObjReg, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return errors.New("failed to get %s", bootstrap.NodeObjRegistry)
	}

	// load dependencies
	registry, ok := nodeObjReg[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("failed to get %s", documents.BootstrappedRegistry)
	}

	configService, ok := nodeObjReg[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	InvoiceUnpaidService, ok := nodeObjReg[bootstrap.BootstrappedInvoiceUnpaid].(nft.InvoiceUnpaid)
	if !ok {
		return errors.New("failed to get %s", bootstrap.BootstrappedInvoiceUnpaid)
	}

	// register documents (common)
	documentpb.RegisterDocumentServiceServer(grpcServer, documents.GRPCHandler(configService, registry))
	err := documentpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// register document types
	err = registerDocumentTypes(ctx, nodeObjReg, grpcServer, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// register other api endpoints
	err = registerAPIs(ctx, cfg, InvoiceUnpaidService, configService, nodeObjReg, grpcServer, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	return nil
}

func registerAPIs(ctx context.Context, cfg Config, InvoiceUnpaidService nft.InvoiceUnpaid, configService config.Service, nodeObjReg map[string]interface{}, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {

	// healthcheck
	hcCfg := cfg.(healthcheck.Config)
	healthpb.RegisterHealthCheckServiceServer(grpcServer, healthcheck.GRPCHandler(hcCfg))
	err := healthpb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// nft api
	nftpb.RegisterNFTServiceServer(grpcServer, nft.GRPCHandler(configService, InvoiceUnpaidService))
	err = nftpb.RegisterNFTServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// account api
	accountpb.RegisterAccountServiceServer(grpcServer, configstore.GRPCAccountHandler(configService))
	err = accountpb.RegisterAccountServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// transactions
	txSrv := nodeObjReg[jobs.BootstrappedService].(jobs.Manager)
	h := jobsv1.GRPCHandler(txSrv, configService)
	jobspb.RegisterJobServiceServer(grpcServer, h)
	return jobspb.RegisterJobServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
}

func registerDocumentTypes(ctx context.Context, nodeObjReg map[string]interface{}, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// register invoice
	invHandler, ok := nodeObjReg[invoice.BootstrappedInvoiceHandler].(invoicepb.DocumentServiceServer)
	if !ok {
		return errors.New("invoice grpc handler not registered")
	}

	invoicepb.RegisterDocumentServiceServer(grpcServer, invHandler)
	err := invoicepb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// register purchase order
	poHandler, ok := nodeObjReg[purchaseorder.BootstrappedPOHandler].(purchaseorderpb.DocumentServiceServer)
	if !ok {
		return errors.New("purchase order grpc handler not registered")
	}

	purchaseorderpb.RegisterDocumentServiceServer(grpcServer, poHandler)
	err = purchaseorderpb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// register entity
	entityHandler, ok := nodeObjReg[entity.BootstrappedEntityHandler].(entitypb.DocumentServiceServer)
	if !ok {
		return errors.New("entity grpc handler not registered")
	}

	entitypb.RegisterDocumentServiceServer(grpcServer, entityHandler)
	return entitypb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
}
