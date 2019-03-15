package anchors

import (
	"context"
	"time"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchorRepository")

// AnchorRepository defines a set of functions that can be
// implemented by any type that stores and retrieves the anchoring, and pre anchoring details.
type AnchorRepository interface {
	PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (confirmations chan bool, err error)
	CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, documentProofs [][32]byte) (chan bool, error)
	GetAnchor(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error)
	HasValidPreCommit(anchorID AnchorID) bool
}
