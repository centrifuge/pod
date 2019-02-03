package nft

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/utils"
)

// TokenIDLength is the length of an NFT token ID
const TokenIDLength = 32

// TokenID is uint256 in Solidity (256 bits | max value is 2^256-1)
// tokenID should be random 32 bytes (32 byte = 256 bits)
type TokenID [TokenIDLength]byte

// NewTokenID returns a new random TokenID
func NewTokenID() TokenID {
	var tid [TokenIDLength]byte
	copy(tid[:], utils.RandomSlice(TokenIDLength))
	return tid
}

func FromString(hexStr string) (TokenID, error) {
	tokenIDBytes, err := hexutil.Decode(hexStr)
	if err != nil {
		return NewTokenID(), err
	}
	if len(tokenIDBytes) != TokenIDLength {
		return NewTokenID(), errors.New("the provided hex string doesn't match the TokenID representation length")
	}
	var tid [TokenIDLength]byte
	copy(tid[:], tokenIDBytes)
	return tid, nil
}

// BigInt converts tokenID to big int
func (t TokenID) BigInt() *big.Int {
	return utils.ByteSliceToBigInt(t[:])
}

// URI gets the URI for this token
func (t TokenID) URI() string {
	// TODO please fix this
	return "http:=//www.centrifuge.io/DUMMY_URI_SERVICE"
}

func (t TokenID) String() string {
	return hexutil.Encode(t[:])
}

// PaymentObligation handles transactions related to minting of NFTs
type PaymentObligation interface {
	// MintNFT mints an NFT
	MintNFT(ctx context.Context, documentID []byte, registryAddress, depositAddress string, proofFields []string) (*MintNFTResponse, chan bool, error)
}

// MintNFTResponse holds tokenID and transaction ID.
type MintNFTResponse struct {
	TokenID       string
	TransactionID string
}
