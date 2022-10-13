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

// GetNFTOwner retrieves the owner of the NFT provided in the request.
func (s *Service) GetNFTOwner(collectionID types.U64, itemID types.U128) (*types.AccountID, error) {
	return s.nftSrvV3.GetNFTOwner(collectionID, itemID)
}

// CreateNFTCollection creates the NFT collection provided in the request.
func (s *Service) CreateNFTCollection(ctx context.Context, collectionID types.U64) (*nftv3.CreateNFTCollectionResponse, error) {
	return s.nftSrvV3.CreateNFTCollection(ctx, collectionID)
}

// ItemMetadataOfNFT retrieves the metadata of an NFT item.
func (s *Service) ItemMetadataOfNFT(collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error) {
	return s.nftSrvV3.GetItemMetadata(collectionID, itemID)
}

// ItemAttributeOfNFT retrieves an attribute of an NFT item.
func (s *Service) ItemAttributeOfNFT(collectionID types.U64, itemID types.U128, key string) ([]byte, error) {
	return s.nftSrvV3.GetItemAttribute(collectionID, itemID, key)
}
