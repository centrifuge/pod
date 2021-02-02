package nft

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// TokenIDLength is the length of an NFT token ID
	TokenIDLength = 32

	// LowEntropyTokenIDMax is the max of a low entropy NFT token ID big integer. Used only for special cases.
	LowEntropyTokenIDMax = "999999999999999"
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

// NewLowEntropyTokenID returns a new low entropy(less than LowEntropyTokenIDMax) TokenID.
// The Dharma NFT Collateralizer and other contracts require tokenIds that are shorter than
// the ERC721 standard bytes32. This option reduces the maximum integer of the tokenId to 999,999,999,999,999.
// There are security implications of doing this. Specifically the risk of two users picking the
// same token id and minting it at the same time goes up and it theoretically could lead to a loss of an
// NFT with large enough NFTRegistries (>100'000 tokens). It is not recommended to use this option.
// TODO(ved): not valid anymore. remove this
func NewLowEntropyTokenID() TokenID {
	var tid [TokenIDLength]byte
	// error is ignored here because the input is a constant.
	n, _ := utils.RandomBigInt(LowEntropyTokenIDMax)
	nByt := n.Bytes()
	// prefix with zeroes to match the bigendian big integer bytes for smart contract
	copy(tid[:], append(make([]byte, TokenIDLength-len(nByt)), nByt...))
	return tid
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

// Service defines the NFT service to mint and transfer NFTs.
type Service interface {
	// MintNFT mints an NFT
	MintNFT(ctx context.Context, request MintNFTRequest) (*TokenResponse, chan error, error)
	// TransferFrom transfers an NFT to another address
	TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID TokenID) (*TokenResponse, chan error, error)
	// OwnerOf returns the owner of an NFT
	OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error)
}

// TokenResponse holds tokenID and transaction ID.
type TokenResponse struct {
	TokenID string
	JobID   string
}
