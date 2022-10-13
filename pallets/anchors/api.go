package anchors

import (
	"context"
	"time"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("anchor_api")
)

const (
	ErrAnchorRetrieval   = errors.Error("couldn't retrieve anchor")
	ErrEmptyDocumentRoot = errors.Error("document root is empty")
)

const (
	// preCommit is centrifuge chain module function name for pre-commit call.
	preCommit = "Anchor.pre_commit"

	// commit is centrifuge chain module function name for commit call.
	commit = "Anchor.commit"

	// getByID is centrifuge chain module function name for getAnchorByID call
	getByID = "anchor_getAnchorById"
)

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

// API defines a set of functions that can be
// implemented by any type that stores and retrieves the anchoring, and pre anchoring details.
type API interface {

	// PreCommitAnchor takes a lock to write the next anchorID using signingRoot as a proof
	PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (err error)

	// CommitAnchor commits the document with given anchorID. If there is a precommit for this anchorID,
	// proof is used to verify before committing the anchor
	CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) error

	// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
	GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error)
}

type api struct {
	anchorLifeSpan time.Duration

	cfgService config.Service
	centAPI    centchain.API
	proxyAPI   proxy.API
}

func NewAPI(
	anchorLifeSpan time.Duration,
	cfgService config.Service,
	centAPI centchain.API,
	proxyAPI proxy.API,
) API {
	return &api{
		anchorLifeSpan,
		cfgService,
		centAPI,
		proxyAPI,
	}
}

// AnchorData holds data returned from previously anchored data against centchain
type AnchorData struct {
	AnchorID     types.Hash `json:"id"`
	DocumentRoot types.Hash `json:"doc_root"`
	BlockNumber  uint32     `json:"anchored_block"`
}

// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
// Returns a nil error when the anchor data is found else returns a non nil error
func (a *api) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	var ad AnchorData
	h := types.NewHash(anchorID[:])
	err = a.centAPI.Call(&ad, getByID, h)

	if err != nil {
		log.Errorf("Couldn't retrieve anchor: %s", err)

		return docRoot, anchoredTime, ErrAnchorRetrieval
	}

	if utils.IsEmptyByte32(ad.DocumentRoot) {
		log.Errorf("Document root is empty, anchor id: %s", anchorID.String())

		return docRoot, anchoredTime, ErrEmptyDocumentRoot
	}

	// TODO(ved): return the anchored time
	return DocumentRoot(ad.DocumentRoot), time.Unix(0, 0), nil
}

// PreCommitAnchor will call the transaction PreCommit substrate module
func (a *api) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		log.Errorf("Couldn't retrieve account from context: %s", err)

		return errors.ErrContextAccountRetrieval
	}

	meta, err := a.centAPI.GetMetadataLatest()
	if err != nil {
		log.Errorf("Couldn't retrieve metadata: %s", err)

		return errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(meta, preCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return errors.ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		log.Errorf("Couldn't retrieve pod operator: %s", err)

		return errors.ErrPodOperatorRetrieval
	}

	_, err = a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

	if err != nil {
		log.Errorf("Couldn't execute proxy call: %s", err)

		return errors.ErrProxyCall
	}

	return nil
}

// CommitAnchor will send a commit transaction to CentChain.
func (a *api) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		log.Errorf("Couldn't retrieve account from context: %s", err)

		return errors.ErrContextAccountRetrieval
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve metadata: %s", err)

		return errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		commit,
		types.NewHash(anchorID[:]),
		types.NewHash(documentRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(time.Now().UTC().Add(a.anchorLifeSpan)),
	)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return errors.ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		log.Errorf("Couldn't retrieve pod operator: %s", err)

		return errors.ErrPodOperatorRetrieval
	}

	_, err = a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

	if err != nil {
		log.Errorf("Couldn't execute proxy call: %s", err)

		return errors.ErrProxyCall
	}

	return nil
}
