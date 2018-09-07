package anchoring

import (
	"math/big"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// wrapper for the Ethereum implementation
type AnchorRepository interface {
	PreCommitAnchor(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error)
	CommitAnchor(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error)
}

func PreCommitAnchor(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.PreCommitAnchor(anchorId, signingRoot, centrifugeId, signature, expirationBlock)
	if err != nil {
		log.Errorf("Failed to pre-commit the anchor [id:%x, hash:%x ]: %v", anchorId, signingRoot, err)
	}
	return confirmations, err
}

func CommitAnchor(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.CommitAnchor(anchorId, documentRoot, centrifugeId, documentProofs, signature)
	if err != nil {
		log.Errorf("Failed to commit the anchor [id:%x, hash:%x ]: %v", anchorId, documentRoot, err)
	}
	return confirmations, err
}

// getConfiguredRepository will later pull a configured repository (if not only using Ethereum as the anchor repository)
// For now hard-coded to the Ethereum setup
func getConfiguredRepository() (AnchorRepository, error) {
	return &EthereumAnchorRepository{}, nil
}
