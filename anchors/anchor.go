package anchors

import (
	"errors"
	"math/big"

	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// AnchorIDLength is the length in bytes of the AnchorID
	AnchorIDLength = 32

	// DocumentRootLength is the length in bytes of the DocumentRoot
	DocumentRootLength = 32

	// DocumentProofLength is the length in bytes of a single proof
	DocumentProofLength = 32

	// AnchorSchemaVersion as stored on public repository
	AnchorSchemaVersion uint = 1
)

// AnchorID type is byte array of length AnchorIDLength
type AnchorID [AnchorIDLength]byte

// Config defines required functions for the package Anchors
type Config interface {
	GetEthereumDefaultAccountName() string
	GetEthereumContextWaitTimeout() time.Duration
	GetContractAddress(address string) common.Address
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

// RandomDocumentRoot returns a randomly generated DocumentRoot
func RandomDocumentRoot() DocumentRoot {
	root, _ := ToDocumentRoot(utils.RandomSlice(DocumentRootLength))
	return root
}

// PreCommitData holds required document details for pre-commit
type PreCommitData struct {
	AnchorID        AnchorID
	SigningRoot     DocumentRoot
	CentrifugeID    identity.CentID
	Signature       []byte
	ExpirationBlock *big.Int
	SchemaVersion   uint
}

// CommitData holds required document details for anchoring
type CommitData struct {
	BlockHeight    uint64
	AnchorID       AnchorID
	DocumentRoot   DocumentRoot
	CentrifugeID   identity.CentID
	DocumentProofs [][DocumentProofLength]byte
	Signature      []byte
	SchemaVersion  uint
}

// WatchCommit holds the commit data received from ethereum event
type WatchCommit struct {
	CommitData *CommitData
	Error      error
}

// WatchPreCommit holds the pre commit data received from ethereum event
type WatchPreCommit struct {
	PreCommit *PreCommitData
	Error     error
}

// supportedSchemaVersion returns the current AnchorSchemaVersion
func supportedSchemaVersion() uint {
	return AnchorSchemaVersion
}

// newPreCommitData returns a PreCommitData with passed in details
func newPreCommitData(anchorID AnchorID, signingRoot DocumentRoot, centrifugeID identity.CentID, signature []byte, expirationBlock *big.Int) (preCommitData *PreCommitData) {
	return &PreCommitData{
		AnchorID:        anchorID,
		SigningRoot:     signingRoot,
		CentrifugeID:    centrifugeID,
		Signature:       signature,
		ExpirationBlock: expirationBlock,
		SchemaVersion:   supportedSchemaVersion(),
	}
}

// NewCommitData returns a CommitData with passed in details
func NewCommitData(blockHeight uint64, anchorID AnchorID, documentRoot DocumentRoot, centrifugeID identity.CentID, documentProofs [][32]byte, signature []byte) (commitData *CommitData) {
	return &CommitData{
		BlockHeight:    blockHeight,
		AnchorID:       anchorID,
		DocumentRoot:   documentRoot,
		CentrifugeID:   centrifugeID,
		DocumentProofs: documentProofs,
		Signature:      signature,
	}
}

// GenerateCommitHash generates Keccak256 message from AnchorID, CentID, DocumentRoot
func GenerateCommitHash(anchorID AnchorID, centrifugeID identity.CentID, documentRoot DocumentRoot) []byte {
	msg := append(anchorID[:], documentRoot[:]...)
	msg = append(msg, centrifugeID[:]...)
	return crypto.Keccak256(msg)
}
