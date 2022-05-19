package v3

import (
	"context"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

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

// OwnerOfNFT retrieves the owner of the NFT provided in the request.
func (s *Service) OwnerOfNFT(ctx context.Context, req *nftv3.OwnerOfRequest) (*nftv3.OwnerOfResponse, error) {
	return s.nftSrvV3.OwnerOf(ctx, req)
}

// CreateNFTClass creates the NFT class provided in the request.
func (s *Service) CreateNFTClass(ctx context.Context, req *nftv3.CreateNFTClassRequest) (*nftv3.CreateNFTClassResponse, error) {
	return s.nftSrvV3.CreateNFTClass(ctx, req)
}

// InstanceMetadataOfNFT retrieves the metadata of an NFT instance.
func (s *Service) InstanceMetadataOfNFT(ctx context.Context, req *nftv3.InstanceMetadataOfRequest) (*types.InstanceMetadata, error) {
	return s.nftSrvV3.InstanceMetadataOf(ctx, req)
}
