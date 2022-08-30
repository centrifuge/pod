package anchors

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	// preCommit is centrifuge chain module function name for pre-commit call.
	preCommit = "Anchor.pre_commit"

	// commit is centrifuge chain module function name for commit call.
	commit = "Anchor.commit"

	// getByID is centrifuge chain module function name for getAnchorByID call
	getByID = "anchor_getAnchorById"
)

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

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
	log *logging.ZapEventLogger

	anchorLifeSpan time.Duration

	cfgService config.Service
	api        centchain.API
	proxyAPI   proxy.API
}

func newService(
	anchorLifeSpan time.Duration,
	cfgService config.Service,
	api centchain.API,
	proxyAPI proxy.API,
) Service {
	log := logging.Logger("anchor_service")

	return &service{
		log,
		anchorLifeSpan,
		cfgService,
		api,
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
func (s *service) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	var ad AnchorData
	h := types.NewHash(anchorID[:])
	err = s.api.Call(&ad, getByID, h)

	if err != nil {
		s.log.Errorf("Couldn't retrieve anchor: %s", err)

		return docRoot, anchoredTime, ErrAnchorRetrieval
	}

	if utils.IsEmptyByte32(ad.DocumentRoot) {
		s.log.Errorf("Document root is empty, anchor id: %s", anchorID.String())

		return docRoot, anchoredTime, ErrEmptyDocumentRoot
	}

	// TODO(ved): return the anchored time
	return DocumentRoot(ad.DocumentRoot), time.Unix(0, 0), nil
}

// PreCommitAnchor will call the transaction PreCommit substrate module
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return errors.ErrContextAccountRetrieval
	}

	meta, err := s.api.GetMetadataLatest()
	if err != nil {
		s.log.Errorf("Couldn't retrieve metadata: %s", err)

		return errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(meta, preCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))

	if err != nil {
		s.log.Errorf("Couldn't create call: %s", err)

		return errors.ErrCallCreation
	}

	podOperator, err := s.cfgService.GetPodOperator()

	if err != nil {
		s.log.Errorf("Couldn't retrieve pod operator: %s", err)

		return errors.ErrPodOperatorRetrieval
	}

	_, err = s.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(types.PodOperation),
		call,
	)

	if err != nil {
		s.log.Errorf("Couldn't execute proxy call: %s", err)

		return errors.ErrProxyCall
	}

	return nil
}

// CommitAnchor will send a commit transaction to CentChain.
func (s *service) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return errors.ErrContextAccountRetrieval
	}

	meta, err := s.api.GetMetadataLatest()

	if err != nil {
		s.log.Errorf("Couldn't retrieve metadata: %s", err)

		return errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		commit,
		types.NewHash(anchorID[:]),
		types.NewHash(documentRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(time.Now().UTC().Add(s.anchorLifeSpan)),
	)

	if err != nil {
		s.log.Errorf("Couldn't create call: %s", err)

		return errors.ErrCallCreation
	}

	podOperator, err := s.cfgService.GetPodOperator()

	if err != nil {
		s.log.Errorf("Couldn't retrieve pod operator: %s", err)

		return errors.ErrPodOperatorRetrieval
	}

	_, err = s.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(types.PodOperation),
		call,
	)

	if err != nil {
		s.log.Errorf("Couldn't execute proxy call: %s", err)

		return errors.ErrProxyCall
	}

	return nil
}
