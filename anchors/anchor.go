package anchors

import (
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	GetEthereumContextWaitTimeout() time.Duration
	GetEthereumGasLimit(op config.ContractOp) uint64
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

// RandomDocumentRoot returns a randomly generated DocumentRoot
func RandomDocumentRoot() DocumentRoot {
	root, _ := ToDocumentRoot(utils.RandomSlice(DocumentRootLength))
	return root
}

// PreCommitData holds required document details for pre-commit
type PreCommitData struct {
	AnchorID      AnchorID
	SigningRoot   DocumentRoot
	SchemaVersion uint
}

// CommitData holds required document details for anchoring
type CommitData struct {
	AnchorID      AnchorID
	DocumentRoot  DocumentRoot
	DocumentProof [DocumentProofLength]byte
	SchemaVersion uint
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
func newPreCommitData(anchorID AnchorID, signingRoot DocumentRoot) (preCommitData *PreCommitData) {
	return &PreCommitData{
		AnchorID:      anchorID,
		SigningRoot:   signingRoot,
		SchemaVersion: supportedSchemaVersion(),
	}
}

// NewCommitData returns a CommitData with passed in details
func NewCommitData(anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) (commitData *CommitData) {
	return &CommitData{
		AnchorID:      anchorID,
		DocumentRoot:  documentRoot,
		DocumentProof: proof,
	}
}

// GenerateCommitHash generates Keccak256 message from AnchorID, CentID, DocumentRoot
func GenerateCommitHash(anchorID AnchorID, centrifugeID identity.DID, documentRoot DocumentRoot) []byte {
	msg := append(anchorID[:], documentRoot[:]...)
	msg = append(msg, centrifugeID[:]...)
	return crypto.Keccak256(msg)
}
