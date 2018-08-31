package repository

import "math/big"

type PreCommitData struct {
	anchorId *big.Int
	signingRoot [32]byte
	centrifugeId *big.Int
	signature []byte
	expirationBlock *big.Int
	SchemaVersion uint
}

type CommitData struct {
	anchorId *big.Int
	documentRoot [32]byte
	centrifugeId *big.Int
	documentProofs [][32]byte
	signature []byte
	SchemaVersion uint
}

type WatchCommit struct {
	CommitData *CommitData
	Error  error
}

type WatchPreCommit struct {
	PreCommit *PreCommitData
	Error  error
}

//Supported anchor schema version as stored on public repository
const ANCHOR_SCHEMA_VERSION uint = 1

func SupportedSchemaVersion() uint {
	return ANCHOR_SCHEMA_VERSION
}

func generatePreCommitData(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (preCommitData *PreCommitData, err error) {
	preCommitData = &PreCommitData{}
	preCommitData.anchorId = anchorId
	preCommitData.signingRoot = signingRoot
	preCommitData.centrifugeId = centrifugeId
	preCommitData.signature = signature
	preCommitData.expirationBlock = expirationBlock
	preCommitData.SchemaVersion = SupportedSchemaVersion()
	return preCommitData, nil
}

func generateCommitData(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (commitData *CommitData, err error) {
	commitData = &CommitData{}
	commitData.anchorId = anchorId
	commitData.documentRoot = documentRoot
	commitData.centrifugeId = centrifugeId
	commitData.documentProofs = documentProofs
	commitData.signature = signature
	return commitData, nil

}


