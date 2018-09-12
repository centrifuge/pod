package anchoring

import (
	"math/big"

	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	AnchorIDLength      = 32
	RootLength          = 32
	DocumentProofLength = 32
)

type AnchorID [AnchorIDLength]byte

func NewAnchorID(anchorBytes []byte) (AnchorID, error) {
	var anchorBytesFixed [AnchorIDLength]byte
	if !tools.IsValidByteSliceForLength(anchorBytes, AnchorIDLength) {
		return anchorBytesFixed, errors.New("invalid length byte slice provided for anchorId")
	}
	copy(anchorBytesFixed[:], anchorBytes[:AnchorIDLength])
	return anchorBytesFixed, nil
}

func (a *AnchorID) toBigInt() *big.Int {
	return tools.ByteSliceToBigInt(a[:])
}

type DocRoot [RootLength]byte

func NewDocRoot(docRootBytes []byte) (DocRoot, error) {
	var rootBytes [RootLength]byte
	if !tools.IsValidByteSliceForLength(docRootBytes, RootLength) {
		return rootBytes, errors.New("invalid length byte slice provided for docRoot")
	}
	copy(rootBytes[:], docRootBytes[:RootLength])
	return rootBytes, nil
}

type PreCommitData struct {
	AnchorId        AnchorID
	SigningRoot     DocRoot
	CentrifugeID    identity.CentID
	Signature       []byte
	ExpirationBlock *big.Int
	SchemaVersion   uint
}

type CommitData struct {
	AnchorId       AnchorID
	DocumentRoot   DocRoot
	CentrifugeID   identity.CentID
	DocumentProofs [][32]byte
	Signature      []byte
	SchemaVersion  uint
}

type WatchCommit struct {
	CommitData *CommitData
	Error      error
}

type WatchPreCommit struct {
	PreCommit *PreCommitData
	Error     error
}

//Supported anchor schema version as stored on public repository
const AnchorSchemaVersion uint = 1

func SupportedSchemaVersion() uint {
	return AnchorSchemaVersion
}

func NewPreCommitData(anchorId AnchorID, signingRoot DocRoot, centrifugeID identity.CentID, signature []byte, expirationBlock *big.Int) (preCommitData *PreCommitData) {
	preCommitData = &PreCommitData{}
	preCommitData.AnchorId = anchorId
	preCommitData.SigningRoot = signingRoot
	preCommitData.CentrifugeID = centrifugeID
	preCommitData.Signature = signature
	preCommitData.ExpirationBlock = expirationBlock
	preCommitData.SchemaVersion = SupportedSchemaVersion()
	return preCommitData
}

func NewCommitData(anchorId AnchorID, documentRoot DocRoot, centrifugeID identity.CentID, documentProofs [][32]byte, signature []byte) (commitData *CommitData) {
	commitData = &CommitData{}
	commitData.AnchorId = anchorId
	commitData.DocumentRoot = documentRoot
	commitData.CentrifugeID = centrifugeID
	commitData.DocumentProofs = documentProofs
	commitData.Signature = signature
	return commitData
}

func GenerateCommitHash(anchorId AnchorID, centrifugeID identity.CentID, documentRoot DocRoot) []byte {
	message := append(anchorId[:], documentRoot[:]...)
	message = append(message, centrifugeID[:]...)
	messageToSign := crypto.Keccak256(message)
	return messageToSign
}
