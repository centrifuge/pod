package api

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/nft"
	funpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// registerServices registers all endpoints to the grpc server
func registerServices(ctx context.Context, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	// node object registry
	nodeObjReg, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return errors.New("failed to get %s", bootstrap.NodeObjRegistry)
	}

	configService, ok := nodeObjReg[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	invoiceUnpaidService, ok := nodeObjReg[bootstrap.BootstrappedInvoiceUnpaid].(nft.InvoiceUnpaid)
	if !ok {
		return errors.New("failed to get %s", bootstrap.BootstrappedInvoiceUnpaid)
	}

	// register document types
	err := registerDocumentTypes(ctx, nodeObjReg, grpcServer, gwmux, addr, dopts)
	if err != nil {
		return err
	}

	// nft api
	nftpb.RegisterNFTServiceServer(grpcServer, nft.GRPCHandler(configService, invoiceUnpaidService))
	return nftpb.RegisterNFTServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
}

func registerDocumentTypes(ctx context.Context, nodeObjReg map[string]interface{}, grpcServer *grpc.Server, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) error {
	fundingHandler, ok := nodeObjReg[funding.BootstrappedFundingAPIHandler].(funpb.FundingServiceServer)
	if !ok {
		return errors.New("funding API handler not registered")
	}

	funpb.RegisterFundingServiceServer(grpcServer, fundingHandler)
	return funpb.RegisterFundingServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
}
