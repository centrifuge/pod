package nft

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Config is an interface to configurations required by nft package
type Config interface {
	GetIdentityID() ([]byte, error)
	GetEthereumDefaultAccountName() string
	GetContractAddress(address string) common.Address
	GetEthereumContextWaitTimeout() time.Duration
}

// PaymentObligation handles transactions related to minting of NFTs
type PaymentObligation interface {

	// MintNFT mints an NFT
	MintNFT(documentID []byte, registryAddress, depositAddress string, proofFields []string) (<-chan *watchTokenMinted, error)
}

type watchTokenMinted struct {
	TokenID *big.Int
	Err     error
}
