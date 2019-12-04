package anchors

import (
	"context"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/types"

	"github.com/centrifuge/go-centrifuge/centchain"
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
}

func newService(config Config, anchorRepository Repository, queue *queue.Server, client centchain.API, jobsMan jobs.Manager) AnchorRepository {
	return &service{config: config, anchorRepository: anchorRepository, client: client, queue: queue, jobsMan: jobsMan}
}

// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
// Returns a nil error when the anchor data is found else returns a non nil error
func (s *service) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	r, err := s.anchorRepository.GetAnchorByID(anchorID.BigInt())
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
	bts, err := types.HexDecodeString(r.DocumentRoot.Hex())
	if err != nil {
		return docRoot, anchoredTime, err
	}
	dr, err := ToDocumentRoot(bts)
	if err != nil {
		return docRoot, anchoredTime, err
	}

	return dr, time.Unix(0, 0), nil
}

// PreCommitAnchor will call the transaction PreCommit substrate module
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (confirmations chan error, err error) {
	return s.anchorRepository.PreCommit(ctx, anchorID, signingRoot)
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
	return s.anchorRepository.Commit(ctx, anchorID, documentRoot, proof, time.Now().UTC().Add(s.config.GetCentChainAnchorLifespan()))
}
