package nft

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("nft-api")

type grpcHandler struct {
	config  config.Service
	service InvoiceUnpaid
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(config config.Service, InvoiceUnpaid InvoiceUnpaid) nftpb.NFTServiceServer {
	return &grpcHandler{config: config, service: InvoiceUnpaid}
}

// MintInvoiceUnpaidNFT will be called from the client API to mint an NFT out of an unpaid invoice
func (g grpcHandler) MintInvoiceUnpaidNFT(ctx context.Context, request *nftpb.NFTMintInvoiceUnpaidRequest) (*nftpb.NFTMintResponse, error) {
	apiLog.Infof("Received request to Mint an Invoice Unpaid NFT for invoice %s and deposit address %s", request.DocumentId, request.DepositAddress)
	ctxHeader, err := contextutil.Context(ctx, g.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	// Get proof fields
	proofFields, err := g.service.GetRequiredInvoiceUnpaidProofFields(ctxHeader)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	cfg, err := g.config.GetConfig()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}
	poRegistry := cfg.GetContractAddress(config.InvoiceUnpaidNFT)

	identifier, err := hexutil.Decode(request.DocumentId)
	if err != nil {
		return nil, err
	}

	req := MintNFTRequest{
		DocumentID:               identifier,
		RegistryAddress:          poRegistry,
		DepositAddress:           common.HexToAddress(request.DepositAddress),
		ProofFields:              proofFields,
		GrantNFTReadAccess:       true,
		SubmitNFTReadAccessProof: true,
		SubmitTokenProof:         true,
	}

	resp, _, err := g.service.MintNFT(ctxHeader, req)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &nftpb.NFTMintResponse{
		Header: &nftpb.ResponseHeader{JobId: resp.JobID},
	}, nil
}
