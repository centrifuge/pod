package nft

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("nft-api")

type grpcHandler struct {
	service PaymentObligation
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(payOb PaymentObligation) nftpb.NFTServiceServer {
	return &grpcHandler{service: payOb}
}

// MintNFT will be called from the client API to mint an NFT
func (g grpcHandler) MintNFT(context context.Context, request *nftpb.NFTMintRequest) (*nftpb.NFTMintResponse, error) {
	apiLog.Infof("Received request to Mint an NFT with  %s with proof fields %s", request.Identifier, request.ProofFields)

	err := validateParameters(request)
	if err != nil {
		return nil, err
	}
	identifier, err := hexutil.Decode(request.Identifier)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}

	confirmation, err := g.service.MintNFT(identifier, request.RegistryAddress, request.DepositAddress, request.ProofFields)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}
	watchToken := <-confirmation
	return &nftpb.NFTMintResponse{TokenId: watchToken.TokenID.String()}, watchToken.Err
}

func validateParameters(request *nftpb.NFTMintRequest) error {

	if !common.IsHexAddress(request.RegistryAddress) {
		return centerrors.New(code.Unknown, "RegistryAddress is not a valid Ethereum address")
	}

	if !common.IsHexAddress(request.DepositAddress) {
		return centerrors.New(code.Unknown, "DepositAddress is not a valid Ethereum address")
	}

	return nil

}
