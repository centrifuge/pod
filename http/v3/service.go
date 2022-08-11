package v3

import (
	"context"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Service is the entry point for all the V3 APIs.
type Service struct {
	nftSrvV3 nftv3.Service
}

// MintNFT mints an NFT for the document provided in the request.
func (s *Service) MintNFT(ctx context.Context, req *nftv3.MintNFTRequest, documentPending bool) (*nftv3.MintNFTResponse, error) {
	return s.nftSrvV3.MintNFT(ctx, req, documentPending)
}

// OwnerOfNFT retrieves the owner of the NFT provided in the request.
func (s *Service) OwnerOfNFT(ctx context.Context, req *nftv3.OwnerOfRequest) (*nftv3.OwnerOfResponse, error) {
	return s.nftSrvV3.OwnerOf(ctx, req)
}

// CreateNFTClass creates the NFT collection provided in the request.
func (s *Service) CreateNFTClass(ctx context.Context, req *nftv3.CreateNFTCollectionRequest) (*nftv3.CreateNFTCollectionResponse, error) {
	return s.nftSrvV3.CreateNFTCollection(ctx, req)
}

// ItemMetadataOfNFT retrieves the metadata of an NFT item.
func (s *Service) ItemMetadataOfNFT(ctx context.Context, req *nftv3.GetItemMetadataRequest) (*types.ItemMetadata, error) {
	return s.nftSrvV3.GetItemMetadata(ctx, req)
}
