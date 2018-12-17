package nft

import "math/big"

// PaymentObligation handles transactions related to minting of NFTs
type PaymentObligation interface {

	// MintNFT mints an NFT
	MintNFT(documentID []byte, registryAddress, depositAddress string, proofFields []string) (<-chan *watchTokenMinted, error)
}

type watchTokenMinted struct {
	TokenID *big.Int
	Err     error
}
