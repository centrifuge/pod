package nft

// PaymentObligation handles transactions related to minting of NFTs
type PaymentObligation interface {

	// MintNFT mints an NFT
	MintNFT(documentID []byte, docType, registryAddress, depositAddress string, proofFields []string) (string, error)
}
