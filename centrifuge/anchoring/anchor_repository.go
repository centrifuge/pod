package anchoring

import (
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// wrapper for the Ethereum implementation
type AnchorRepository interface {
	PreCommitAnchor(anchorId AnchorID, signingRoot DocRoot, centrifugeID identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error)
	CommitAnchor(anchorId AnchorID, documentRoot DocRoot, centrifugeId identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error)
}

func PreCommitAnchor(anchorId AnchorID, signingRoot DocRoot, centrifugeId identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.PreCommitAnchor(anchorId, signingRoot, centrifugeId, signature, expirationBlock)
	if err != nil {
		log.Errorf("Failed to pre-commit the anchor [id:%x, hash:%x ]: %v", anchorId, signingRoot, err)
	}
	return confirmations, err
}

func CommitAnchor(anchorId AnchorID, documentRoot DocRoot, centrifugeID identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.CommitAnchor(anchorId, documentRoot, centrifugeID, documentProofs, signature)
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
