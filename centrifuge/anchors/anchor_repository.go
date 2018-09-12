package anchors

import (
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// wrapper for the Ethereum implementation
type AnchorRepository interface {
	PreCommitAnchor(anchorID AnchorID, signingRoot DocRoot, centrifugeID identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error)
	CommitAnchor(anchorID AnchorID, documentRoot DocRoot, centrifugeId identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error)
}

func PreCommitAnchor(anchorID AnchorID, signingRoot DocRoot, centrifugeId identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.PreCommitAnchor(anchorID, signingRoot, centrifugeId, signature, expirationBlock)
	if err != nil {
		log.Errorf("Failed to pre-commit the anchor [id:%x, hash:%x ]: %v", anchorID, signingRoot, err)
	}
	return confirmations, err
}

func CommitAnchor(anchorID AnchorID, documentRoot DocRoot, centrifugeID identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.CommitAnchor(anchorID, documentRoot, centrifugeID, documentProofs, signature)
	if err != nil {
		log.Errorf("Failed to commit the anchor [id:%x, hash:%x ]: %v", anchorID, documentRoot, err)
	}
	return confirmations, err
}

// getConfiguredRepository will later pull a configured repository (if not only using Ethereum as the anchor repository)
// For now hard-coded to the Ethereum setup
func getConfiguredRepository() (AnchorRepository, error) {
	return &EthereumAnchorRepository{}, nil
}
