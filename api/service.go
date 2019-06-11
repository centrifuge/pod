package api

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/extension/funding"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	funpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	invoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	purchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/go-chi/chi"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// registerServices registers all endpoints to the grpc server
func registerServices(ctx context.Context, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) (mux *chi.Mux, err error) {
	// node object registry
	nodeObjReg, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to get %s", bootstrap.NodeObjRegistry)
	}

	configService, ok := nodeObjReg[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return nil, errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	invoiceUnpaidService, ok := nodeObjReg[bootstrap.BootstrappedInvoiceUnpaid].(nft.InvoiceUnpaid)
	if !ok {
		return nil, errors.New("failed to get %s", bootstrap.BootstrappedInvoiceUnpaid)
	}

	docService, ok := nodeObjReg[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return nil, errors.New("failed to get %s", documents.BootstrappedDocumentService)
	}

	// register document types
	err = registerDocumentTypes(ctx, nodeObjReg, grpcServer, gwmux, addr, dopts)
	if err != nil {
		return nil, err
	}

	// register other api endpoints
	err = registerAPIs(ctx, invoiceUnpaidService, configService, grpcServer, gwmux, addr, dopts)
	if err != nil {
		return nil, err
	}

	cfg := nodeObjReg[bootstrap.BootstrappedConfig].(config.Configuration)
	jobsMan := nodeObjReg[jobs.BootstrappedService].(jobs.Manager)
	return httpapi.Router(cfg, configService, invoiceUnpaidService, docService, jobsMan), nil
}

func registerAPIs(ctx context.Context, InvoiceUnpaidService nft.InvoiceUnpaid, configService config.Service, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {

	// nft api
	nftpb.RegisterNFTServiceServer(grpcServer, nft.GRPCHandler(configService, InvoiceUnpaidService))
	err := nftpb.RegisterNFTServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// account api
	accountpb.RegisterAccountServiceServer(grpcServer, configstore.GRPCAccountHandler(configService))
	return accountpb.RegisterAccountServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
}

func registerDocumentTypes(ctx context.Context, nodeObjReg map[string]interface{}, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// register invoice
	invHandler, ok := nodeObjReg[invoice.BootstrappedInvoiceHandler].(invoicepb.InvoiceServiceServer)
	if !ok {
		return errors.New("invoice grpc handler not registered")
	}

	invoicepb.RegisterInvoiceServiceServer(grpcServer, invHandler)
	err := invoicepb.RegisterInvoiceServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// register purchase order
	poHandler, ok := nodeObjReg[purchaseorder.BootstrappedPOHandler].(purchaseorderpb.PurchaseOrderServiceServer)
	if !ok {
		return errors.New("purchase order grpc handler not registered")
	}

	purchaseorderpb.RegisterPurchaseOrderServiceServer(grpcServer, poHandler)
	err = purchaseorderpb.RegisterPurchaseOrderServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// register entity
	entityHandler, ok := nodeObjReg[entity.BootstrappedEntityHandler].(entitypb.EntityServiceServer)
	if !ok {
		return errors.New("entity grpc handler not registered")
	}

	entitypb.RegisterEntityServiceServer(grpcServer, entityHandler)
	err = entitypb.RegisterEntityServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	fundingHandler, ok := nodeObjReg[funding.BootstrappedFundingAPIHandler].(funpb.FundingServiceServer)
	if !ok {
		return errors.New("funding API handler not registered")
	}

	funpb.RegisterFundingServiceServer(grpcServer, fundingHandler)
	return funpb.RegisterFundingServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
}
