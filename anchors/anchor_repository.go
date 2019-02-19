package anchors

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// AnchorRepository defines a set of functions that can be
// implemented by any type that stores and retrieves the anchoring, and pre anchoring details.
type AnchorRepository interface {
	PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot, centID identity.CentID, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error)
	CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, centID identity.DID, documentProofs [][32]byte, signature []byte) (confirmations <-chan *WatchCommit, err error)
	GetDocumentRootOf(anchorID AnchorID) (DocumentRoot, error)
}
