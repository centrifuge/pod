package nft

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// PaymentObligation handles transactions related to minting of NFTs
type PaymentObligation interface {
	// MintNFT mints an NFT
	MintNFT(ctx context.Context, documentID []byte, registryAddress, depositAddress string, proofFields []string) (*MintNFTResponse, error)
}

// MintNFTResponse holds tokenID and transaction ID.
type MintNFTResponse struct {
	TokenID       string
	TransactionID string
}

// TokenRegistry defines NFT retrieval functions.
type TokenRegistry interface {
	// OwnerOf to retrieve owner of the tokenID
	OwnerOf(registry common.Address, tokenID []byte) (common.Address, error)
}
