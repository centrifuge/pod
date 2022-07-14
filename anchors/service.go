package anchors

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	// preCommit is centrifuge chain module function name for pre-commit call.
	preCommit = "Anchor.pre_commit"

	// commit is centrifuge chain module function name for commit call.
	commit = "Anchor.commit"

	// getByID is centrifuge chain module function name for getAnchorByID call
	getByID = "anchor_getAnchorById"
)

// Service defines a set of functions that can be
// implemented by any type that stores and retrieves the anchoring, and pre anchoring details.
type Service interface {

	// PreCommitAnchor takes a lock to write the next anchorID using signingRoot as a proof
	PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (err error)

	// CommitAnchor commits the document with given anchorID. If there is a precommit for this anchorID,
	// proof is used to verify before committing the anchor
	CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) error

	// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
	GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error)
}

type service struct {
	config Config
	api    centchain.API
}

func newService(config Config, api centchain.API) Service {
	return &service{config: config, api: api}
}

// AnchorData holds data returned from previously anchored data against centchain
type AnchorData struct {
	AnchorID     types.Hash `json:"id"`
	DocumentRoot types.Hash `json:"doc_root"`
	BlockNumber  uint32     `json:"anchored_block"`
}

// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
// Returns a nil error when the anchor data is found else returns a non nil error
func (s *service) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	var ad AnchorData
	h := types.NewHash(anchorID[:])
	err = s.api.Call(&ad, getByID, h)
	if err != nil {
		return docRoot, anchoredTime, fmt.Errorf("failed to get anchor: %w", err)
	}

	if utils.IsEmptyByte32(ad.DocumentRoot) {
		return docRoot, anchoredTime, errors.New("anchor data empty for id: %v", anchorID.String())
	}

	// TODO(ved): return the anchored time
	return DocumentRoot(ad.DocumentRoot), time.Unix(0, 0), nil
}

// PreCommitAnchor will call the transaction PreCommit substrate module
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	// TODO(cdamian): Create proxy type for anchoring?
	accProxy, err := acc.GetAccountProxies().WithProxyType(types.NFTManagement)
	if err != nil {
		return err
	}

	krp, err := accProxy.ToKeyringPair()

	if err != nil {
		return err
	}

	meta, err := s.api.GetMetadataLatest()
	if err != nil {
		return err
	}

	c, err := types.NewCall(meta, preCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))
	if err != nil {
		return err
	}

	_, err = s.api.SubmitAndWatch(ctx, meta, c, *krp)
	if err != nil {
		return fmt.Errorf("failed to precommit document: %w", err)
	}

	return nil
}

// CommitAnchor will send a commit transaction to CentChain.
func (s *service) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	// TODO(cdamian): Create proxy type for anchoring?
	accProxy, err := acc.GetAccountProxies().WithProxyType(types.NFTManagement)
	if err != nil {
		return err
	}

	krp, err := accProxy.ToKeyringPair()

	if err != nil {
		return err
	}

	meta, err := s.api.GetMetadataLatest()
	if err != nil {
		return err
	}

	c, err := types.NewCall(
		meta,
		commit,
		types.NewHash(anchorID[:]),
		types.NewHash(documentRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(time.Now().UTC().Add(s.config.GetCentChainAnchorLifespan())))
	if err != nil {
		return err
	}

	_, err = s.api.SubmitAndWatch(ctx, meta, c, *krp)
	if err != nil {
		return fmt.Errorf("failed to commit document: %w", err)
	}

	return nil
}
