package anchors

import (
	"context"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

const (
	// PreCommit is centrifuge chain module function name for pre-commit call.
	PreCommit = "Anchor.pre_commit"

	// Commit is centrifuge chain module function name for commit call.
	Commit = "Anchor.commit"
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

	GetAnchorById(id *big.Int) (struct {
		AnchorId     *big.Int
		DocumentRoot [32]byte
		BlockNumber  uint32
	}, error)

	HasValidPreCommit(anchorId *big.Int) (bool, error)
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

	c, err := types.NewCall(meta, PreCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))
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
		Commit,
		types.NewHash(anchorIDPreImage[:]),
		types.NewHash(documentRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(storedUntil))
	if err != nil {
		return txHash, bn, sig, err
	}

	return r.api.SubmitExtrinsic(meta, c, krp)
}

func (r repository) HasValidPreCommit(anchorId *big.Int) (bool, error) {
	return false, errors.New("not implemented")
}

func (r repository) GetAnchorById(id *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	BlockNumber  uint32
}, error) {
	return struct {
		AnchorId     *big.Int
		DocumentRoot [32]byte
		BlockNumber  uint32
	}{}, errors.New("not implemented")
}
