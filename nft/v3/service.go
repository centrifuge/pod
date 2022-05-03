package v3

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
	"go.uber.org/zap"
)

// MintNFTRequest request to mint nft on centrifuge chain.
type MintNFTRequest struct {
	DocumentID []byte
	PublicInfo []string // save to IPFS
	ClassID    types.U64
	Owner      types.AccountID // substrate account ID
}

// MintNFTResponse holds tokenID and transaction ID.
type MintNFTResponse struct {
	JobID      string
	InstanceID types.U128
}

type Service interface {
	MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error)
}

type service struct {
	docSrv     documents.Service
	dispatcher jobs.Dispatcher
	api        UniquesAPI
	log        *logging.ZapEventLogger
}

func newService(
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
	api UniquesAPI,
) Service {
	log := logging.Logger("nft_v2")
	return &service{
		docSrv,
		dispatcher,
		api,
		log,
	}
}

const (
	// TODO(cdamian): Drop the hardcoded class ID when we know how to proceed with this.
	defaultClassID types.U64 = 1234
)

func (s *service) MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error) {
	log := s.log.With(
		zap.ByteString("doc_id", req.DocumentID),
		zap.Uint64("class_id", uint64(req.ClassID)),
		zap.Any("owner", req.Owner),
	)

	tc, err := contextutil.Account(ctx)
	if err != nil {
		log.Error("Couldn't retrieve account from context", zap.Error(err))

		return nil, err
	}

	// TODO(cdamian): Remove overwrite.
	req.ClassID = defaultClassID

	if err := s.validateDocNFTs(ctx, req); err != nil {
		log.Error("Document NFT validation failed", zap.Error(err))

		return nil, err
	}

	instanceID, err := s.generateInstanceID(ctx, req.ClassID)

	if err != nil {
		log.Error("Couldn't generate instance ID", zap.Error(err))

		return nil, err
	}

	didBytes := tc.GetIdentityID()

	did, err := identity.NewDIDFromBytes(didBytes)

	if err != nil {
		log.Error("Couldn't generate identity", zap.Error(err))

		return nil, err
	}

	jobID, err := s.dispatchNFTMintJob(did, instanceID, req)

	if err != nil {
		log.Error("Couldn't dispatch NFT mint job", zap.Error(err))

		return nil, err
	}

	return &MintNFTResponse{
		JobID:      jobID.Hex(),
		InstanceID: instanceID,
	}, nil
}

func (s *service) validateDocNFTs(ctx context.Context, req *MintNFTRequest) error {
	log := s.log.With(
		zap.ByteString("doc_id", req.DocumentID),
		zap.Uint64("class_id", uint64(req.ClassID)),
		zap.Any("owner", req.Owner),
	)

	doc, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)

	if err != nil {
		log.Error("Couldn't get current doc version", zap.Error(err))

		return fmt.Errorf("couldn't get current document version: %w", err)
	}

	if len(doc.NFTs()) == 0 {
		log.Info("Document has no NFTs, proceeding")

		return nil
	}

	for _, nft := range doc.NFTs() {
		var nftClassID types.U64

		if err := types.DecodeFromBytes(nft.ClassId, &nftClassID); err != nil {
			log.Error("Couldn't decode document class ID", zap.Error(err))

			return err
		}

		if nftClassID != req.ClassID {
			continue
		}

		var instanceID types.U128

		if err := types.DecodeFromBytes(nft.InstanceId, &instanceID); err != nil {
			log.Error("Couldn't decode instance ID", zap.Error(err))

			return err
		}

		log = log.With(zap.Int64("instance_id", instanceID.Int64()))

		instanceDetails, err := s.api.GetInstanceDetails(ctx, nftClassID, instanceID)

		if err != nil {
			log.Error("Couldn't get instance details", zap.Error(err))

			return err
		}

		if instanceDetails == nil {
			log.Info("NFT instance found but not minted")

			return nil
		}

		anchorID, err := anchors.ToAnchorID(doc.CurrentVersion())

		if err != nil {
			log.Error("Couldn't get anchor ID", zap.Error(err))

			return err
		}

		return fmt.Errorf("instance with ID %d was already minted for doc with anchor %s", instanceID, anchorID)
	}

	return nil
}

func (s *service) dispatchNFTMintJob(did identity.DID, instanceID types.U128, req *MintNFTRequest) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Mint NFT on Centrifuge Chain",
		mintNFTV3Job,
		"add_nft_v3_to_document",
		[]interface{}{
			did,
			instanceID,
			req,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if _, err := s.dispatcher.Dispatch(did, job); err != nil {
		s.log.Error("Couldn't dispatch mint NFT job", zap.Error(err))

		return nil, fmt.Errorf("failed to dispatch mint NFT job: %w", err)
	}

	return job.ID, nil
}

func (s *service) generateInstanceID(ctx context.Context, classID types.U64) (types.U128, error) {
	var instanceID types.U128

	for {
		select {
		case <-ctx.Done():
			return instanceID, ctx.Err()
		default:
			instanceID = types.NewU128(*big.NewInt(int64(rand.Int())))

			instanceDetails, err := s.api.GetInstanceDetails(ctx, classID, instanceID)

			if err != nil {
				return instanceID, fmt.Errorf("couldn't get instance details: %w", err)
			}

			if instanceDetails == nil {
				return instanceID, nil
			}
		}
	}
}
