package v3

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/centrifuge/go-centrifuge/validation"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
)

type Service interface {
	CreateNFTClass(ctx context.Context, req *CreateNFTClassRequest) (*CreateNFTClassResponse, error)
	MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error)
	OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error)
	InstanceMetadataOf(ctx context.Context, req *ItemMetadataOfRequest) (*types.InstanceMetadata, error)
}

type service struct {
	log *logging.ZapEventLogger

	docSrv     documents.Service
	dispatcher jobs.Dispatcher
	api        UniquesAPI
}

func NewService(
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
	api UniquesAPI,
) Service {
	log := logging.Logger("nft_v3")
	return &service{
		log,
		docSrv,
		dispatcher,
		api,
	}
}

func (s *service) MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error) {
	if err := validation.Validate(validation.NewValidator(req, mintNFTRequestValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, ErrRequestInvalid
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrAccountFromContextRetrieval
	}

	if err := s.validateDocNFTs(ctx, req); err != nil {
		s.log.Errorf("Document NFT validation failed: %s", err)

		return nil, err
	}

	instanceID, err := s.generateInstanceID(ctx, req.ClassID)

	if err != nil {
		s.log.Errorf("Couldn't generate instance ID: %s", err)

		return nil, ErrItemIDGeneration
	}

	jobID, err := s.dispatchNFTMintJob(acc.GetIdentity(), instanceID, req)

	if err != nil {
		s.log.Errorf("Couldn't dispatch NFT mint job: %s", err)

		return nil, ErrMintJobDispatch
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

		return ErrDocumentRetrieval
	}

	if len(doc.CcNfts()) == 0 {
		s.log.Info("Document has no NFTs, proceeding")

		return nil
	}

	for _, nft := range doc.CcNfts() {
		var nftClassID types.U64

		if err := types.DecodeFromBytes(nft.ClassId, &nftClassID); err != nil {
			s.log.Errorf("Couldn't decode document class ID: %s", err)

			return ErrCollectionIDDecoding
		}

		if nftClassID != req.ClassID {
			continue
		}

		var instanceID types.U128

		if err := types.DecodeFromBytes(nft.InstanceId, &instanceID); err != nil {
			s.log.Errorf("Couldn't decode instance ID: %s", err)

			return ErrItemIDDecoding
		}

		_, err := s.api.GetItemDetails(ctx, nftClassID, instanceID)

		if err != nil {
			if errors.Is(err, ErrItemDetailsNotFound) {
				s.log.Info("NFT instance found but not minted")

				return nil
			}

			s.log.Errorf("Couldn't get instance details: %s", err)

			return ErrItemDetailsRetrieval
		}

		docVersion := doc.CurrentVersion()

		anchorID, err := anchors.ToAnchorID(docVersion)

		if err != nil {
			s.log.Errorf("Couldn't parse anchor ID for doc with version %s: %s", docVersion, err)
		}

		err = ErrItemAlreadyMinted

		return fmt.Errorf("instance ID %d was already minted for doc with anchor %s: %w", instanceID, anchorID, err)

	}

	return nil
}

func (s *service) dispatchNFTMintJob(identity *types.AccountID, instanceID types.U128, req *MintNFTRequest) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Mint NFT on Centrifuge Chain",
		mintNFTV3Job,
		"add_nft_v3_to_document",
		[]interface{}{
			identity,
			instanceID,
			req,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if err := s.dispatchJob(identity, job); err != nil {
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

			_, err := s.api.GetItemDetails(ctx, classID, instanceID)

			if err != nil {
				if errors.Is(err, ErrItemDetailsNotFound) {
					return instanceID, nil
				}

				return instanceID, err
			}
		}
	}
}

func (s *service) OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error) {
	if err := validation.Validate(validation.NewValidator(req, ownerOfValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, ErrRequestInvalid
	}

	instanceDetails, err := s.api.GetItemDetails(ctx, req.ClassID, req.InstanceID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve the instance details: %s", err)

		if errors.Is(err, ErrItemDetailsNotFound) {
			return nil, ErrOwnerNotFound
		}

		return nil, ErrItemDetailsRetrieval
	}

	return &OwnerOfResponse{
		ClassID:    req.ClassID,
		InstanceID: req.InstanceID,
		AccountID:  instanceDetails.Owner,
	}, nil
}

func (s *service) CreateNFTClass(ctx context.Context, req *CreateNFTClassRequest) (*CreateNFTClassResponse, error) {
	if err := validation.Validate(validation.NewValidator(req, createNFTClassRequestValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, ErrRequestInvalid
	}

	acc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrAccountFromContextRetrieval
	}

	classExists, err := s.classExists(ctx, req.ClassID)

	if err != nil {
		s.log.Errorf("Couldn't check if class already exists: %s", err)

		return nil, ErrCollectionCheck
	}

	if classExists {
		s.log.Errorf("Class already exists")

		return nil, ErrCollectionAlreadyExists
	}

	jobID, err := s.dispatchCreateClassJob(acc.GetIdentity(), req.ClassID)

	if err != nil {
		s.log.Errorf("Couldn't create class: %s", err)

		return nil, ErrCreateCollectionJobDispatch
	}

	return &CreateNFTClassResponse{
		JobID:   jobID.Hex(),
		ClassID: req.ClassID,
	}, nil
}

func (s *service) classExists(ctx context.Context, classID types.U64) (bool, error) {
	_, err := s.api.GetCollectionDetails(ctx, classID)

	if err != nil {
		if errors.Is(err, ErrCollectionDetailsNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *service) dispatchCreateClassJob(identity *types.AccountID, classID types.U64) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Create NFT class on Centrifuge Chain",
		createNFTClassV3Job,
		"create_nft_class_v3",
		[]interface{}{
			identity,
			classID,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if err := s.dispatchJob(identity, job); err != nil {
		s.log.Errorf("Couldn't dispatch create class job: %s", err)

		return nil, fmt.Errorf("failed to dispatch create class job: %w", err)
	}

	return job.ID, nil
}

func (s *service) dispatchJob(identity *types.AccountID, job *gocelery.Job) error {
	if _, err := s.dispatcher.Dispatch(identity, job); err != nil {
		return err
	}

	return nil
}

func (s *service) InstanceMetadataOf(ctx context.Context, req *ItemMetadataOfRequest) (*types.InstanceMetadata, error) {
	if err := validation.Validate(validation.NewValidator(req, instanceMetadataOfRequestValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, ErrRequestInvalid
	}

	instanceMetadata, err := s.api.GetInstanceMetadata(ctx, req.ClassID, req.InstanceID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve instance metadata: %s", err)

		if errors.Is(err, ErrItemMetadataNotFound) {
			return nil, err
		}

		return nil, ErrItemMetadataRetrieval
	}

	return instanceMetadata, nil
}
