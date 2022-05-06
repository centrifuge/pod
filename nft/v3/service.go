package v3

import (
	"context"
	"encoding/gob"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
)

func init() {
	gob.Register(types.U128{})
	gob.Register(MintNFTRequest{})
}

// OwnerOfRequest is the request object for the retrieval of the owner of an NFT on Centrifuge chain.
type OwnerOfRequest struct {
	ClassID    types.U64
	InstanceID types.U128
}

// OwnerOfResponse is the response object for a OwnerOfRequest, it holds the AccountID of the owner of an NFT.
type OwnerOfResponse struct {
	ClassID    types.U64
	InstanceID types.U128
	AccountID  types.AccountID
}

// MintNFTRequest is the request object for minting an NFT on Centrifuge chain.
type MintNFTRequest struct {
	DocumentID []byte
	PublicInfo []string // save to IPFS
	ClassID    types.U64
	Owner      types.AccountID // substrate account ID
}

// MintNFTResponse is the response object for a MintNFTRequest, it holds the job ID and instance ID of the NFT.
type MintNFTResponse struct {
	JobID      string
	InstanceID types.U128
}

type Service interface {
	MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error)
	OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error)
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
	log := logging.Logger("nft_v3")
	return &service{
		docSrv,
		dispatcher,
		api,
		log,
	}
}

func (s *service) MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, err
	}

	if err := s.validateDocNFTs(ctx, req); err != nil {
		s.log.Errorf("Document NFT validation failed: %s", err)

		return nil, err
	}

	instanceID, err := s.generateInstanceID(ctx, req.ClassID)

	if err != nil {
		s.log.Errorf("Couldn't generate instance ID: %s", err)

		return nil, err
	}

	didBytes := tc.GetIdentityID()

	did, err := identity.NewDIDFromBytes(didBytes)

	if err != nil {
		s.log.Errorf("Couldn't generate identity: %s", err)

		return nil, err
	}

	jobID, err := s.dispatchNFTMintJob(did, instanceID, req)

	if err != nil {
		s.log.Errorf("Couldn't dispatch NFT mint job: %s", err)

		return nil, err
	}

	return &MintNFTResponse{
		JobID:      jobID.Hex(),
		InstanceID: instanceID,
	}, nil
}

func (s *service) validateDocNFTs(ctx context.Context, req *MintNFTRequest) error {
	doc, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)

	if err != nil {
		s.log.Errorf("Couldn't get current doc version: %s", err)

		return fmt.Errorf("couldn't get current document version: %w", err)
	}

	if len(doc.CcNfts()) == 0 {
		s.log.Info("Document has no NFTs, proceeding")

		return nil
	}

	for _, nft := range doc.CcNfts() {
		var nftClassID types.U64

		if err := types.DecodeFromBytes(nft.ClassId, &nftClassID); err != nil {
			s.log.Errorf("Couldn't decode document class ID: %s", err)

			return err
		}

		if nftClassID != req.ClassID {
			continue
		}

		var instanceID types.U128

		if err := types.DecodeFromBytes(nft.InstanceId, &instanceID); err != nil {
			s.log.Errorf("Couldn't decode instance ID: %s", err)

			return err
		}

		instanceDetails, err := s.api.GetInstanceDetails(ctx, nftClassID, instanceID)

		if err != nil {
			s.log.Errorf("Couldn't get instance details: %s", err)

			return err
		}

		if instanceDetails == nil {
			s.log.Info("NFT instance found but not minted")

			return nil
		}

		anchorID, err := anchors.ToAnchorID(doc.CurrentVersion())

		if err != nil {
			s.log.Errorf("Couldn't get anchor ID: %s", err)

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
		s.log.Errorf("Couldn't dispatch mint NFT job: %s", err)

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

const (
	ErrInstanceDetailsNotFound = errors.Error("instance details not found")
)

func (s *service) OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error) {
	instanceDetails, err := s.api.GetInstanceDetails(ctx, req.ClassID, req.InstanceID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve the instance details: %s", err)

		return nil, err
	}

	if instanceDetails == nil {
		s.log.Error("Instance details not found")

		return nil, ErrInstanceDetailsNotFound
	}

	return &OwnerOfResponse{
		ClassID:    req.ClassID,
		InstanceID: req.InstanceID,
		AccountID:  instanceDetails.Owner,
	}, nil
}
