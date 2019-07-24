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

	// PreCommitAnchor will call the transaction PreCommit on the smart contract, to pre commit a document update
	PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (confirmations chan error, err error)

	// CommitAnchor will send a commit transaction to Ethereum.
	CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) (chan error, error)

	// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
	GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error)

	// HasValidPreCommit checks if the given anchorID has a valid pre-commit
	HasValidPreCommit(anchorID AnchorID) bool
}
