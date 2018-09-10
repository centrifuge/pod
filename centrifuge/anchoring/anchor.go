package anchoring

import (
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	AnchorIdLength      = 32
	RootLength          = 32
	DocumentProofLength = 32
)

type AnchorId [AnchorIdLength]byte

func NewAnchorId(anchorBytes []byte) AnchorId {
	var bytes [AnchorIdLength]byte
	copy(bytes[:], anchorBytes[:AnchorIdLength])
	return bytes
}

func (a *AnchorId) toBigInt() *big.Int {
	return tools.ByteSliceToBigInt(a[:])
}

type DocRoot [RootLength]byte

func NewDocRoot(docRootBytes []byte) DocRoot {
	var bytes [RootLength]byte
	copy(bytes[:], docRootBytes[:RootLength])
	return bytes
}

type PreCommitData struct {
	AnchorId        AnchorId
	SigningRoot     DocRoot
	CentrifugeId    [identity.CentIdByteLength]byte
	Signature       []byte
	ExpirationBlock *big.Int
	SchemaVersion   uint
}

type CommitData struct {
	AnchorId       AnchorId
	DocumentRoot   DocRoot
	CentrifugeId   [identity.CentIdByteLength]byte
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

func NewPreCommitData(anchorId AnchorId, signingRoot DocRoot, centrifugeId [identity.CentIdByteLength]byte, signature []byte, expirationBlock *big.Int) (preCommitData *PreCommitData) {
	preCommitData = &PreCommitData{}
	preCommitData.AnchorId = anchorId
	preCommitData.SigningRoot = signingRoot
	preCommitData.CentrifugeId = centrifugeId
	preCommitData.Signature = signature
	preCommitData.ExpirationBlock = expirationBlock
	preCommitData.SchemaVersion = SupportedSchemaVersion()
	return preCommitData
}

func NewCommitData(anchorId AnchorId, documentRoot DocRoot, centrifugeId [identity.CentIdByteLength]byte, documentProofs [][32]byte, signature []byte) (commitData *CommitData) {
	commitData = &CommitData{}
	commitData.AnchorId = anchorId
	commitData.DocumentRoot = documentRoot
	commitData.CentrifugeId = centrifugeId
	commitData.DocumentProofs = documentProofs
	commitData.Signature = signature
	return commitData
}

func GenerateCommitHash(anchorId AnchorId, centrifugeId [identity.CentIdByteLength]byte, documentRoot DocRoot) []byte {
	message := append(anchorId[:], documentRoot[:]...)
	message = append(message, centrifugeId[:]...)
	messageToSign := crypto.Keccak256(message)
	return messageToSign
}
