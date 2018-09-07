package repository

import "math/big"

type PreCommitData struct {
	AnchorId        *big.Int
	SigningRoot     [32]byte
	CentrifugeId    *big.Int
	Signature       []byte
	ExpirationBlock *big.Int
	SchemaVersion   uint
}

type CommitData struct {
	AnchorId       *big.Int
	DocumentRoot   [32]byte
	CentrifugeId   *big.Int
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

func NewPreCommitData(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (preCommitData *PreCommitData, err error) {
	preCommitData = &PreCommitData{}
	preCommitData.AnchorId = anchorId
	preCommitData.SigningRoot = signingRoot
	preCommitData.CentrifugeId = centrifugeId
	preCommitData.Signature = signature
	preCommitData.ExpirationBlock = expirationBlock
	preCommitData.SchemaVersion = SupportedSchemaVersion()
	return preCommitData, nil
}

func NewCommitData(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (commitData *CommitData, err error) {
	commitData = &CommitData{}
	commitData.AnchorId = anchorId
	commitData.DocumentRoot = documentRoot
	commitData.CentrifugeId = centrifugeId
	commitData.DocumentProofs = documentProofs
	commitData.Signature = signature
	return commitData, nil

}
