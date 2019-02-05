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
	service PaymentObligation
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(config config.Service, payOb PaymentObligation) nftpb.NFTServiceServer {
	return &grpcHandler{config: config, service: payOb}
}

// MintNFT will be called from the client API to mint an NFT
func (g grpcHandler) MintNFT(ctx context.Context, request *nftpb.NFTMintRequest) (*nftpb.NFTMintResponse, error) {
	apiLog.Infof("Received request to Mint an NFT with  %s with proof fields %s", request.Identifier, request.ProofFields)
	ctxHeader, err := contextutil.Context(ctx, g.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	err = validateParameters(request)
	if err != nil {
		return nil, err
	}

	identifier, err := hexutil.Decode(request.Identifier)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	resp, _, err := g.service.MintNFT(ctxHeader, identifier, request.RegistryAddress, request.DepositAddress, request.ProofFields)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &nftpb.NFTMintResponse{
		TokenId: resp.TokenID,
		Header:  &nftpb.ResponseHeader{TransactionId: resp.TransactionID},
	}, nil
}

func validateParameters(request *nftpb.NFTMintRequest) error {
	if !common.IsHexAddress(request.RegistryAddress) {
		return centerrors.New(code.Unknown, "registryAddress is not a valid Ethereum address")
	}

	if !common.IsHexAddress(request.DepositAddress) {
		return centerrors.New(code.Unknown, "DepositAddress is not a valid Ethereum address")
	}
	return nil

}
