package nft

import "math/big"

// PaymentObligation handles transactions related to minting of NFTs
type PaymentObligation interface {

	// MintNFT mints an NFT
	MintNFT(documentID []byte, docType, registryAddress, depositAddress string, proofFields []string) (<-chan *WatchTokenMinted, error)
}

type WatchTokenMinted struct {
	TokenID *big.Int
	Err     error
}
