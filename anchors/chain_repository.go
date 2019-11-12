package anchors

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

// Repository defines required APIs to interact with Anchor Repository on Centrifuge Chain.
type Repository interface {
	// PreCommit takes anchorID and signingRoot and submits an extrinsic to the Cent chain.
	// Returns the latest block number before submission and signature attached to the extrinsic.
	PreCommit(
		ctx context.Context,
		anchorID AnchorID,
		signingRoot DocumentRoot) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error)

	// Commit takes anchorID pre image, document root, and proof if pre-commit is submitted for this commit to commit an anchor
	// on chain.
	// Returns latest block number before extrinsic submission and signature attached with the sub
	Commit(
		ctx context.Context,
		anchorIDPreImage AnchorID,
		documentRoot DocumentRoot,
		proof [32]byte,
		storedUntil time.Time) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error)
}

type repository struct {
	api centchain.API
}

// NewRepository returns a new Anchor repository.
func NewRepository(api centchain.API) Repository {
	return repository{api: api}
}

func (r repository) PreCommit(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return txHash, bn, sig, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return txHash, bn, sig, err
	}

	meta, err := r.api.GetMetadataLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	c, err := types.NewCall(meta, "Anchor.pre_commit", types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))
	if err != nil {
		return txHash, bn, sig, err
	}

	return r.api.SubmitExtrinsic(meta, c, krp)
}

func (r repository) Commit(
	ctx context.Context,
	anchorIDPreImage AnchorID,
	documentRoot DocumentRoot,
	proof [32]byte, storedUntil time.Time) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error) {

	acc, err := contextutil.Account(ctx)
	if err != nil {
		return txHash, bn, sig, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return txHash, bn, sig, err
	}

	meta, err := r.api.GetMetadataLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	c, err := types.NewCall(
		meta,
		"Anchor.commit",
		types.NewHash(anchorIDPreImage[:]),
		types.NewHash(documentRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(storedUntil))
	if err != nil {
		return txHash, bn, sig, err
	}

	return r.api.SubmitExtrinsic(meta, c, krp)
}
