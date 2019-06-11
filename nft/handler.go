package nft

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
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
	apiLog.Infof("Received request to Mint an Invoice Unpaid NFT for invoice %s and deposit address %s", request.Identifier, request.DepositAddress)
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

	identifier, err := hexutil.Decode(request.Identifier)
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

// TokenTransfer will be called to transfer a token owned by the identity contract
func (g grpcHandler) TokenTransfer(ctx context.Context, request *nftpb.TokenTransferRequest) (*nftpb.TokenTransferResponse, error) {
	ctxHeader, err := contextutil.Context(ctx, g.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	tokenID, err := TokenIDFromString(request.TokenId)
	if err != nil {
		return nil, errors.NewTypedError(ErrInvalidParameter, err)
	}

	if !common.IsHexAddress(request.RegistryAddress) {
		return nil, ErrInvalidAddress
	}
	registry := common.HexToAddress(request.RegistryAddress)
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(request.To) {
		return nil, ErrInvalidAddress
	}
	to := common.HexToAddress(request.To)
	if err != nil {
		return nil, err
	}

	resp, _, err := g.service.TransferFrom(ctxHeader, registry, to, tokenID)
	if err != nil {
		return nil, errors.NewTypedError(ErrTokenTransfer, err)
	}
	return &nftpb.TokenTransferResponse{
		Header: &nftpb.ResponseHeader{JobId: resp.JobID},
	}, nil
}

// OwnerOf returns the owner of an NFT
func (g grpcHandler) OwnerOf(ctx context.Context, request *nftpb.OwnerOfRequest) (*nftpb.OwnerOfResponse, error) {
	tokenID, err := TokenIDFromString(request.TokenId)
	if err != nil {
		return nil, errors.NewTypedError(ErrInvalidParameter, err)
	}

	if !common.IsHexAddress(request.RegistryAddress) {
		return nil, ErrInvalidAddress
	}
	registry := common.HexToAddress(request.RegistryAddress)
	if err != nil {
		return nil, err
	}

	owner, err := g.service.OwnerOf(registry, tokenID[:])
	if err != nil {
		return nil, errors.NewTypedError(ErrOwnerOf, err)
	}

	return &nftpb.OwnerOfResponse{TokenId: request.TokenId, Owner: owner.Hex(), RegistryAddress: request.RegistryAddress}, nil
}
