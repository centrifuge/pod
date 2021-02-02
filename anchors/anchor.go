package anchors

import (
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// AnchorIDLength is the length in bytes of the AnchorID
	AnchorIDLength = 32

	// DocumentRootLength is the length in bytes of the DocumentRoot
	DocumentRootLength = 32
)

// AnchorID type is byte array of length AnchorIDLength
type AnchorID [AnchorIDLength]byte

// Config defines required functions for the package Anchors
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
	GetEthereumGasLimit(op config.ContractOp) uint64
	GetCentChainAnchorLifespan() time.Duration
}

// ToAnchorID convert the bytes into AnchorID type
// returns an error if the bytes length != AnchorIDLength
func ToAnchorID(bytes []byte) (AnchorID, error) {
	var id [AnchorIDLength]byte
	if !utils.IsValidByteSliceForLength(bytes, AnchorIDLength) {
		return id, errors.New("invalid length byte slice provided for anchorID")
	}

	copy(id[:], bytes[:AnchorIDLength])
	return id, nil
}

// BigInt returns anchorID in bigInt form
func (a *AnchorID) BigInt() *big.Int {
	return utils.ByteSliceToBigInt(a[:])
}

// String returns anchorID in string form
func (a *AnchorID) String() string {
	return hexutil.Encode(a[:])
}

// DocumentRoot type is byte array of length DocumentRootLength
type DocumentRoot [DocumentRootLength]byte

// ToDocumentRoot converts bytes to DocumentRoot
// returns error if the bytes length != DocumentRootLength
func ToDocumentRoot(bytes []byte) (DocumentRoot, error) {
	var root [DocumentRootLength]byte
	if !utils.IsValidByteSliceForLength(bytes, DocumentRootLength) {
		return root, errors.New("invalid length byte slice provided for docRoot")
	}

	copy(root[:], bytes[:DocumentRootLength])
	return root, nil
}
