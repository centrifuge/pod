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
	//Deprecated old version
	PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot, centID identity.DID, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error)
	CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, documentProofs [][32]byte) (chan bool, error)
	GetDocumentRootOf(anchorID AnchorID) (DocumentRoot, error)
}
