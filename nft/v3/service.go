package v3

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	nodeErrors "github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft/v3/uniques"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
)

type Service interface {
	CreateNFTCollection(ctx context.Context, req *CreateNFTCollectionRequest) (*CreateNFTCollectionResponse, error)
	MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error)
	OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error)
	GetItemMetadata(ctx context.Context, req *GetItemMetadataRequest) (*types.ItemMetadata, error)
}

type service struct {
	log *logging.ZapEventLogger

	docSrv     documents.Service
	dispatcher jobs.Dispatcher
	api        uniques.API
}

func NewService(
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
	api uniques.API,
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

		return nil, nodeErrors.ErrRequestInvalid
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, nodeErrors.ErrContextAccountRetrieval
	}

	if err := s.validateDocNFTs(ctx, req); err != nil {
		s.log.Errorf("Document NFT validation failed: %s", err)

		return nil, err
	}

	itemID, err := s.generateItemID(ctx, req.CollectionID)

	if err != nil {
		s.log.Errorf("Couldn't generate item ID: %s", err)

		return nil, ErrItemIDGeneration
	}

	jobID, err := s.dispatchNFTMintJob(acc, itemID, req)

	if err != nil {
		s.log.Errorf("Couldn't dispatch NFT mint job: %s", err)

		return nil, ErrMintJobDispatch
	}

	return &MintNFTResponse{
		JobID:  jobID.Hex(),
		ItemID: itemID,
	}, nil
}

func (s *service) validateDocNFTs(ctx context.Context, req *MintNFTRequest) error {
	doc, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)

	if err != nil {
		s.log.Errorf("Couldn't get current doc version: %s", err)

		return ErrDocumentRetrieval
	}

	if len(doc.NFTs()) == 0 {
		s.log.Info("Document has no NFTs, proceeding")

		return nil
	}

	for _, nft := range doc.NFTs() {
		var nftCollectionID types.U64

		if err := types.Decode(nft.GetRegistryId(), &nftCollectionID); err != nil {
			s.log.Errorf("Couldn't decode collection ID: %s", err)

			return ErrCollectionIDDecoding
		}

		if nftCollectionID != req.CollectionID {
			continue
		}

		var nftItemID types.U128

		if err := types.Decode(nft.GetTokenId(), &nftItemID); err != nil {
			s.log.Errorf("Couldn't decode item ID: %s", err)

			return ErrItemIDDecoding
		}

		_, err := s.api.GetItemDetails(ctx, nftCollectionID, nftItemID)

		if err != nil {
			if errors.Is(err, uniques.ErrItemDetailsNotFound) {
				s.log.Info("NFT item found but not minted")

				return nil
			}

			s.log.Errorf("Couldn't get instance details: %s", err)

			return err
		}

		docVersion := doc.CurrentVersion()

		anchorID, err := anchors.ToAnchorID(docVersion)

		if err != nil {
			s.log.Warnf("Couldn't parse anchor ID for doc with version %s: %s", docVersion, err)
		}

		err = ErrItemAlreadyMinted

		return fmt.Errorf("instance ID %d was already minted for doc with anchor %s: %w", nftItemID, anchorID, err)

	}

	return nil
}

func (s *service) dispatchNFTMintJob(account config.Account, itemID types.U128, req *MintNFTRequest) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Mint NFT on Centrifuge Chain",
		mintNFTV3Job,
		"add_nft_v3_to_document",
		[]interface{}{
			account,
			itemID,
			req,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if err := s.dispatchJob(account.GetIdentity(), job); err != nil {
		s.log.Errorf("Couldn't dispatch mint NFT job: %s", err)

		return nil, fmt.Errorf("failed to dispatch mint NFT job: %w", err)
	}

	return job.ID, nil
}

func (s *service) generateItemID(ctx context.Context, collectionID types.U64) (types.U128, error) {
	var itemID types.U128

	for {
		select {
		case <-ctx.Done():
			return itemID, ctx.Err()
		default:
			itemID = types.NewU128(*big.NewInt(int64(rand.Int())))

			_, err := s.api.GetItemDetails(ctx, collectionID, itemID)

			if err != nil {
				if errors.Is(err, uniques.ErrItemDetailsNotFound) {
					return itemID, nil
				}

				return itemID, err
			}
		}
	}
}

func (s *service) OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error) {
	if err := validation.Validate(validation.NewValidator(req, ownerOfValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, nodeErrors.ErrRequestInvalid
	}

	instanceDetails, err := s.api.GetItemDetails(ctx, req.CollectionID, req.ItemID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve the instance details: %s", err)

		if errors.Is(err, uniques.ErrItemDetailsNotFound) {
			return nil, ErrOwnerNotFound
		}

		return nil, err
	}

	return &OwnerOfResponse{
		CollectionID: req.CollectionID,
		ItemID:       req.ItemID,
		AccountID:    &instanceDetails.Owner,
	}, nil
}

func (s *service) CreateNFTCollection(ctx context.Context, req *CreateNFTCollectionRequest) (*CreateNFTCollectionResponse, error) {
	if err := validation.Validate(validation.NewValidator(req, createNFTCollectionRequestValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, nodeErrors.ErrRequestInvalid
	}

	acc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, nodeErrors.ErrContextAccountRetrieval
	}

	collectionExists, err := s.collectionExists(ctx, req.CollectionID)

	if err != nil {
		s.log.Errorf("Couldn't check if class already exists: %s", err)

		return nil, ErrCollectionCheck
	}

	if collectionExists {
		s.log.Errorf("Class already exists")

		return nil, ErrCollectionAlreadyExists
	}

	jobID, err := s.dispatchCreateCollectionJob(acc, req.CollectionID)

	if err != nil {
		s.log.Errorf("Couldn't create collection: %s", err)

		return nil, ErrCreateCollectionJobDispatch
	}

	return &CreateNFTCollectionResponse{
		JobID:        jobID.Hex(),
		CollectionID: req.CollectionID,
	}, nil
}

func (s *service) collectionExists(ctx context.Context, classID types.U64) (bool, error) {
	_, err := s.api.GetCollectionDetails(ctx, classID)

	if err != nil {
		if errors.Is(err, uniques.ErrCollectionDetailsNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *service) dispatchCreateCollectionJob(account config.Account, collectionID types.U64) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Create NFT class on Centrifuge Chain",
		createNFTClassV3Job,
		"create_nft_class_v3",
		[]interface{}{
			account,
			collectionID,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if err := s.dispatchJob(account.GetIdentity(), job); err != nil {
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

func (s *service) GetItemMetadata(ctx context.Context, req *GetItemMetadataRequest) (*types.ItemMetadata, error) {
	if err := validation.Validate(validation.NewValidator(req, itemMetadataOfRequestValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, nodeErrors.ErrRequestInvalid
	}

	itemMetadata, err := s.api.GetItemMetadata(ctx, req.CollectionID, req.ItemID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve instance metadata: %s", err)

		if errors.Is(err, uniques.ErrItemMetadataNotFound) {
			return nil, err
		}

		return nil, err
	}

	return itemMetadata, nil
}
