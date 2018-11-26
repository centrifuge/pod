package anchors

import (
	"math/big"

	"github.com/centrifuge/go-centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// AnchorRepository defines a set of functions that can be
// implemented by any type that stores and retrieves the anchoring, and pre anchoring details.
type AnchorRepository interface {
	PreCommitAnchor(anchorID AnchorID, signingRoot DocumentRoot, centID identity.CentID, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error)
	CommitAnchor(anchorID AnchorID, documentRoot DocumentRoot, centID identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error)
	GetDocumentRootOf(anchorID AnchorID) (DocumentRoot, error)
}
