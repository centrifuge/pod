package anchors

import (
	"context"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/jobs"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

const (
	// PreCommit is centrifuge chain module function name for pre-commit call.
	PreCommit = "Anchor.pre_commit"

	// Commit is centrifuge chain module function name for commit call.
	Commit = "Anchor.commit"

	// GetByID is centrifuge chain module function name for getAnchorByID call
	GetByID = "anchor_getAnchorById"
)

// Repository defines required APIs to interact with Anchor Repository on Centrifuge Chain.
type Repository interface {
	// PreCommit takes anchorID and signingRoot and submits an extrinsic to the Cent chain.
	// Returns the confirmation channel.
	PreCommit(
		ctx context.Context,
		anchorID AnchorID,
		signingRoot DocumentRoot) (confirmations chan error, err error)

	// Commit takes anchorID pre image, document root, and proof if pre-commit is submitted for this commit to commit an anchor
	// on chain.
	// Returns confirmations channel
	Commit(
		ctx context.Context,
		anchorIDPreImage AnchorID,
		documentRoot DocumentRoot,
		proof [32]byte,
		storedUntil time.Time) (confirmations chan error, err error)

	// GetAnchorByID returns the anchor stored on-chain
	GetAnchorByID(id *big.Int) (struct {
		AnchorID     *big.Int
		DocumentRoot [32]byte
		BlockNumber  uint32
	}, error)
}

type repository struct {
	api     centchain.API
	jobsMan jobs.Manager
}

// NewRepository returns a new Anchor repository.
func NewRepository(api centchain.API, jobsMan jobs.Manager) Repository {
	return repository{
		api:     api,
		jobsMan: jobsMan,
	}
}

func (r repository) PreCommit(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (confirmations chan error, err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return nil, err
	}

	meta, err := r.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	c, err := types.NewCall(meta, PreCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))
	if err != nil {
		return nil, err
	}

	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	jobID := contextutil.Job(ctx)
	cctx := contextutil.Copy(ctx)
	_, done, err := r.jobsMan.ExecuteWithinJob(cctx, did, jobID, "Check Job for anchor pre-commit", r.api.SubmitAndWatch(cctx, meta, c, krp))

	return done, err
}

func (r repository) Commit(
	ctx context.Context,
	anchorIDPreImage AnchorID,
	documentRoot DocumentRoot,
	proof [32]byte, storedUntil time.Time) (confirmations chan error, err error) {

	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return nil, err
	}

	meta, err := r.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	c, err := types.NewCall(
		meta,
		Commit,
		types.NewHash(anchorIDPreImage[:]),
		types.NewHash(documentRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(storedUntil))
	if err != nil {
		return nil, err
	}

	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	jobID := contextutil.Job(ctx)
	cctx := contextutil.Copy(ctx)
	_, done, err := r.jobsMan.ExecuteWithinJob(cctx, did, jobID, "Check Job for anchor commit", r.api.SubmitAndWatch(cctx, meta, c, krp))

	return done, err
}

func (r repository) GetAnchorByID(id *big.Int) (struct {
	AnchorID     *big.Int
	DocumentRoot [32]byte
	BlockNumber  uint32
}, error) {

	var anchorData struct {
		AnchorID     *big.Int
		DocumentRoot [32]byte
		BlockNumber  uint32
	}

	err := r.api.Call(&anchorData, GetByID, types.NewHash(id.Bytes()))
	if err != nil {
		return anchorData, err
	}
	return anchorData, nil
}
