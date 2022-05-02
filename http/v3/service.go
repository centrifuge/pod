package v3

import (
	"context"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
)

// Service is the entry point for all the V3 APIs.
type Service struct {
	nftSrvV3 nftv3.Service
}

// MintNFT mints an NFT for the document provided in the request.
func (s *Service) MintNFT(ctx context.Context, req *nftv3.MintNFTRequest) (*nftv3.MintNFTResponse, error) {
	return s.nftSrvV3.MintNFT(ctx, req)
}
