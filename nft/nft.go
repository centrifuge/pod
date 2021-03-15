package nft

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// TokenIDLength is the length of an NFT token ID
	TokenIDLength = 32
)

// TokenID is uint256 in Solidity (256 bits | max value is 2^256-1)
// tokenID should be random 32 bytes (32 byte = 256 bits)
type TokenID [TokenIDLength]byte

// NewTokenID returns a new random TokenID
func NewTokenID() TokenID {
	var tid [TokenIDLength]byte
	copy(tid[:], utils.RandomSlice(TokenIDLength))
	return tid
}

// MarshalText converts the Token to its text form
func (t TokenID) MarshalText() (text []byte, err error) {
	return []byte(hexutil.Encode(t[:])), nil
}

// UnmarshalText converts text to TokenID
func (t *TokenID) UnmarshalText(text []byte) error {
	tid, err := TokenIDFromString(string(text))
	if err != nil {
		return err
	}
	*t = tid
	return nil
}

// TokenIDFromString converts given hex string to a TokenID
func TokenIDFromString(hexStr string) (TokenID, error) {
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

func (t TokenID) String() string {
	return hexutil.Encode(t[:])
}

// MintNFTRequest holds required fields for minting NFT
type MintNFTRequest struct {
	DocumentID               []byte
	ProofFields              []string
	RegistryAddress          common.Address
	DepositAddress           common.Address
	AssetManagerAddress      common.Address
	GrantNFTReadAccess       bool
	SubmitTokenProof         bool
	SubmitNFTReadAccessProof bool
}

// MintNFTOnCCRequest request to mint nft on centrifuge chain.
type MintNFTOnCCRequest struct {
	DocumentID         []byte
	ProofFields        []string
	RegistryAddress    common.Address
	DepositAddress     types.AccountID
	GrantNFTReadAccess bool
}

// Service defines the NFT service to mint and transfer NFTs.
type Service interface {
	// MintNFT mints an NFT
	MintNFT(ctx context.Context, request MintNFTRequest) (*TokenResponse, error)
	// TransferFrom transfers an NFT to another address
	TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID TokenID) (*TokenResponse, error)
	// OwnerOf returns the owner of an NFT
	OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error)
	// MintNFTOnCC mints an NFT on Centrifuge chain
	MintNFTOnCC(ctx context.Context, req MintNFTOnCCRequest) (*TokenResponse, error)
	// OwnerOfOnCC returns the owner of the NFT
	OwnerOfOnCC(registry common.Address, tokenID TokenID) (types.AccountID, error)
	// TransferNFT transfers NFT to `to` account
	TransferNFT(ctx context.Context, registry common.Address, tokenID TokenID, to types.AccountID) (*TokenResponse, error)
}

// TokenResponse holds tokenID and transaction ID.
type TokenResponse struct {
	TokenID string
	JobID   string
}
