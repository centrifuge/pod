package anchors

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/transaction"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
)

type service struct {
	config           Config
	anchorRepository Repository
	client           centchain.API
	queue            *queue.Server
	jobsMan          jobs.Manager
	txSvc            transaction.Submitter
}

func newService(config Config, anchorRepository Repository, queue *queue.Server, client centchain.API, jobsMan jobs.Manager, txSvc transaction.Submitter) AnchorRepository {
	return &service{config: config, anchorRepository: anchorRepository, client: client, queue: queue, jobsMan: jobsMan, txSvc: txSvc}
}

// HasValidPreCommit checks if the given anchorID has a valid pre-commit
func (s *service) HasValidPreCommit(anchorID AnchorID) bool {
	r, err := s.anchorRepository.HasValidPreCommit(anchorID.BigInt())
	if err != nil {
		return false
	}
	return r
}

// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
// Returns a nil error when the anchor data is found else returns a non nil error
func (s *service) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	r, err := s.anchorRepository.GetAnchorById(anchorID.BigInt())
	if err != nil {
		return docRoot, anchoredTime, err
	}

	if utils.IsEmptyByte32(r.DocumentRoot) {
		return docRoot, anchoredTime, errors.New("anchor data missing for id: %v", anchorID.String())
	}

	//TODO get block time
	//blk, err := s.client.GetBlockByNumber(context.Background(), big.NewInt(int64(r.BlockNumber)))
	//if err != nil {
	//	return docRoot, anchoredTime, err
	//}

	return r.DocumentRoot, time.Unix(0, 0), nil
}

// PreCommitAnchor will call the transaction PreCommit substrate module
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (confirmations chan error, err error) {
	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	jobID := contextutil.Job(ctx)
	pc := newPreCommitData(anchorID, signingRoot)
	log.Infof("Add Anchor to Pre-commit %s from did:%s", anchorID.String(), did.ToAddress().String())
	_, done, err := s.jobsMan.ExecuteWithinJob(contextutil.Copy(ctx), did, jobID, "Check Job for anchor commit",
		s.txSvc.SubmitAndWatch(s.anchorRepository.PreCommit, pc.AnchorID.BigInt(), pc.SigningRoot))
	if err != nil {
		return nil, err
	}

	return done, nil
}

// getDID returns DID from context.Account
// TODO use did.NewDIDFromContext as soon as IDConfig is deleted
func getDID(ctx context.Context) (identity.DID, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return identity.DID{}, err
	}

	addressByte := tc.GetIdentityID()
	return identity.NewDID(common.BytesToAddress(addressByte)), nil
}

// CommitAnchor will send a commit transaction to CentChain.
func (s *service) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) (chan error, error) {
	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	jobID := contextutil.Job(ctx)

	cd := NewCommitData(anchorID, documentRoot, proof)

	log.Infof("Add Anchor to Commit %s from did:%s", anchorID.String(), did.ToAddress().String())
	_, done, err := s.jobsMan.ExecuteWithinJob(contextutil.Copy(ctx), did, jobID, "Check Job for anchor commit",
		s.txSvc.SubmitAndWatch(s.anchorRepository.Commit, cd.AnchorID.BigInt(), cd.DocumentRoot, cd.DocumentProof))
	if err != nil {
		return nil, err
	}

	return done, nil
}
