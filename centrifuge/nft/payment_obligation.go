package nft

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type WatchMint struct {
	MintRequestData *MintRequestData
	Error      error
}

type MintRequestData struct {
	To common.Address
	TokenId *big.Int
	TokenURI string
	AnchorId *big.Int
	MerkleRoot [32]byte
	Values [3]string
	Salts [3][32]byte
	Proofs [3][][32]byte

}

type PaymentObligation interface {
	Mint(to common.Address, tokenId *big.Int, tokenURI string, anchorId *big.Int, merkleRoot [32]byte, values [3]string, salts [3][32]byte, proofs [3][][32]byte) (<-chan *WatchMint, error)

}
