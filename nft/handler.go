package nft

import (
	"context"

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
func GRPCHandler() nftpb.NFTServiceServer {
	return &grpcHandler{service: GetPaymentObligation()}
}

// MintNFT will be called from the client API to mint an NFT
func (g grpcHandler) MintNFT(context context.Context, request *nftpb.NFTMintRequest) (*nftpb.NFTMintResponse, error) {
	apiLog.Infof("Received request to Mint an NFT for document %s type %s with proof fields %s", request.Identifier, request.Type, request.ProofFields)
	identifier, err := hexutil.Decode(request.Identifier)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}

	confirmation, err := g.service.MintNFT(identifier, request.Type, request.RegistryAddress, request.DepositAddress, request.ProofFields)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}
	watchToken := <-confirmation
	return &nftpb.NFTMintResponse{TokenId: watchToken.TokenID.String()}, watchToken.Err
}
