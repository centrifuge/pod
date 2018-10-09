package nft

import (
	"context"
	nftpb"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/nft"


)

type grpcHandler struct {}


// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler() nftpb.NFTServiceServer {
	return &grpcHandler{}
}


func (g grpcHandler)MintNFT(context.Context, *nftpb.NFTMintRequest) (*nftpb.NFTMintResponse, error) {

	panic("not implemented yet")

}


