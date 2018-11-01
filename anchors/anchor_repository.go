package anchors

import (
	"math/big"

	"github.com/centrifuge/go-centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// AnchorRepository defines a set of functions that can be
// implemented by any type that stores and retrieves the anchoring, and pre anchoring details
type AnchorRepository interface {
	PreCommitAnchor(anchorID AnchorID, signingRoot DocumentRoot, centrifugeID identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error)
	CommitAnchor(anchorID AnchorID, documentRoot DocumentRoot, centrifugeId identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error)
	GetDocumentRootOf(anchorID AnchorID) (DocumentRoot, error)
}

// PreCommitAnchor initiates the PreCommit call on the smart contract
// with passed in variables and returns a channel for transaction confirmation
func PreCommitAnchor(anchorID AnchorID, signingRoot DocumentRoot, centrifugeId identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.PreCommitAnchor(anchorID, signingRoot, centrifugeId, signature, expirationBlock)
	if err != nil {
		log.Errorf("Failed to pre-commit the anchor [id:%x, hash:%x ]: %v", anchorID, signingRoot, err)
	}
	return confirmations, err
}

// CommitAnchor initiates the Commit call on smart contract
// with passed in variables and returns a channel for transaction confirmation
func CommitAnchor(anchorID AnchorID, documentRoot DocumentRoot, centrifugeID identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error) {
	anchorRepository, _ := getConfiguredRepository()

	confirmations, err := anchorRepository.CommitAnchor(anchorID, documentRoot, centrifugeID, documentProofs, signature)
	if err != nil {
		log.Errorf("Failed to commit the anchor [id:%x, hash:%x ]: %v", anchorID, documentRoot, err)
	}
	return confirmations, err
}

// anchorRepository is a singleton to keep track of the anchorRepository
var anchorRepository AnchorRepository

// setAnchorRepository sets the passed in repository as default one
func setAnchorRepository(ar AnchorRepository) {
	anchorRepository = ar
}

// GetAnchorRepository returns default anchor repository
func GetAnchorRepository() AnchorRepository {
	return anchorRepository
}

// getConfiguredRepository will later pull a configured repository (if not only using Ethereum as the anchor repository)
// For now hard-coded to the Ethereum setup
func getConfiguredRepository() (AnchorRepository, error) {
	return anchorRepository, nil
}
