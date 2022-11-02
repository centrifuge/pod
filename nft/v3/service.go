package v3

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	nodeErrors "github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pallets/anchors"
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
)

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

type Service interface {
	CreateNFTCollection(ctx context.Context, collectionID types.U64) (*CreateNFTCollectionResponse, error)
	MintNFT(ctx context.Context, req *MintNFTRequest, documentPending bool) (*MintNFTResponse, error)
	GetNFTOwner(collectionID types.U64, itemID types.U128) (*types.AccountID, error)
	GetItemMetadata(collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error)
	GetItemAttribute(collectionID types.U64, itemID types.U128, key string) ([]byte, error)
}

type service struct {
	log *logging.ZapEventLogger

	pendingDocSrv pending.Service
	docSrv        documents.Service
	dispatcher    jobs.Dispatcher
	api           uniques.API
}

func NewService(
	pendingDocSrv pending.Service,
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
	api uniques.API,
) Service {
	log := logging.Logger("nft_v3_service")
	return &service{
		log,
		pendingDocSrv,
		docSrv,
		dispatcher,
		api,
	}
}

func (s *service) MintNFT(ctx context.Context, req *MintNFTRequest, documentPending bool) (*MintNFTResponse, error) {
	if err := validation.Validate(validation.NewValidator(req, mintNFTRequestValidatorFn)); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, nodeErrors.NewTypedError(nodeErrors.ErrRequestInvalid, err)
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, nodeErrors.ErrContextAccountRetrieval
	}

	if err := s.validateDocNFTs(ctx, req, documentPending); err != nil {
		s.log.Errorf("Document NFT validation failed: %s", err)

		return nil, err
	}

	itemID, err := s.generateItemID(ctx, req.CollectionID)

	if err != nil {
		s.log.Errorf("Couldn't generate item ID: %s", err)

		return nil, ErrItemIDGeneration
	}

	jobID, err := s.dispatchNFTMintJob(acc, itemID, req, documentPending)

	if err != nil {
		s.log.Errorf("Couldn't dispatch NFT mint job: %s", err)

		return nil, ErrMintJobDispatch
	}

	return &MintNFTResponse{
		JobID:  jobID.Hex(),
		ItemID: itemID,
	}, nil
}

func (s *service) CreateNFTCollection(ctx context.Context, collectionID types.U64) (*CreateNFTCollectionResponse, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, nodeErrors.ErrContextAccountRetrieval
	}

	collectionExists, err := s.collectionExists(collectionID)

	if err != nil {
		s.log.Errorf("Couldn't check if collection already exists: %s", err)

		return nil, ErrCollectionCheck
	}

	if collectionExists {
		s.log.Errorf("Collection already exists")

		return nil, ErrCollectionAlreadyExists
	}

	jobID, err := s.dispatchCreateCollectionJob(acc, collectionID)

	if err != nil {
		s.log.Errorf("Couldn't create collection: %s", err)

		return nil, ErrCreateCollectionJobDispatch
	}

	return &CreateNFTCollectionResponse{
		JobID:        jobID.Hex(),
		CollectionID: collectionID,
	}, nil
}

func (s *service) GetItemMetadata(collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error) {
	itemMetadata, err := s.api.GetItemMetadata(collectionID, itemID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve item metadata: %s", err)

		if errors.Is(err, uniques.ErrItemMetadataNotFound) {
			return nil, ErrItemMetadataNotFound
		}

		return nil, ErrItemMetadataRetrieval
	}

	return itemMetadata, nil
}

func (s *service) GetItemAttribute(collectionID types.U64, itemID types.U128, key string) ([]byte, error) {
	value, err := s.api.GetItemAttribute(collectionID, itemID, []byte(key))

	if err != nil {
		s.log.Errorf("Couldn't retrieve item attribute: %s", err)

		if errors.Is(err, uniques.ErrItemAttributeNotFound) {
			return nil, ErrItemAttributeNotFound
		}

		return nil, ErrItemAttributeRetrieval
	}

	return value, nil
}

func (s *service) GetNFTOwner(collectionID types.U64, itemID types.U128) (*types.AccountID, error) {
	itemDetails, err := s.api.GetItemDetails(collectionID, itemID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve the instance details: %s", err)

		if errors.Is(err, uniques.ErrItemDetailsNotFound) {
			return nil, ErrOwnerNotFound
		}

		return nil, ErrOwnerRetrieval
	}

	return &itemDetails.Owner, nil
}

func (s *service) validateDocNFTs(ctx context.Context, req *MintNFTRequest, documentPending bool) error {
	var (
		doc documents.Document
		err error
	)

	if documentPending {
		doc, err = s.pendingDocSrv.Get(ctx, req.DocumentID, documents.Pending)
	} else {
		doc, err = s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	}

	if err != nil {
		s.log.Errorf("Couldn't get retrieve document: %s", err)

		return ErrDocumentRetrieval
	}

	if len(doc.NFTs()) == 0 {
		s.log.Info("Document has no NFTs, proceeding")

		return nil
	}

	for _, nft := range doc.NFTs() {
		var nftCollectionID types.U64

		if err := codec.Decode(nft.GetCollectionId(), &nftCollectionID); err != nil {
			s.log.Errorf("Couldn't decode collection ID: %s", err)

			return ErrCollectionIDDecoding
		}

		if nftCollectionID != req.CollectionID {
			continue
		}

		var nftItemID types.U128

		if err := codec.Decode(nft.GetItemId(), &nftItemID); err != nil {
			s.log.Errorf("Couldn't decode item ID: %s", err)

			return ErrItemIDDecoding
		}

		_, err := s.api.GetItemDetails(nftCollectionID, nftItemID)

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

		return fmt.Errorf("item ID %d was already minted for doc with anchor %s: %w", nftItemID, anchorID, err)
	}

	return nil
}

func (s *service) dispatchNFTMintJob(
	account config.Account,
	itemID types.U128,
	req *MintNFTRequest,
	documentPending bool,
) (gocelery.JobID, error) {
	job := getNFTMintRunnerJob(
		documentPending,
		[]any{
			account,
			itemID,
			req,
		},
	)

	if err := s.dispatchJob(account.GetIdentity(), job); err != nil {
		s.log.Errorf("Couldn't dispatch mint NFT job: %s", err)

		return nil, fmt.Errorf("failed to dispatch mint NFT job: %w", err)
	}

	return job.ID, nil
}

func (s *service) collectionExists(collectionID types.U64) (bool, error) {
	_, err := s.api.GetCollectionDetails(collectionID)

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
		"Create NFT collection on Centrifuge Chain",
		createNFTCollectionV3Job,
		"create_nft_collection_v3",
		[]any{
			account,
			collectionID,
		},
		make(map[string]any),
		time.Time{},
	)

	if err := s.dispatchJob(account.GetIdentity(), job); err != nil {
		s.log.Errorf("Couldn't dispatch create collection job: %s", err)

		return nil, fmt.Errorf("failed to dispatch create collection job: %w", err)
	}

	return job.ID, nil
}

func (s *service) dispatchJob(identity *types.AccountID, job *gocelery.Job) error {
	if _, err := s.dispatcher.Dispatch(identity, job); err != nil {
		return err
	}

	return nil
}

func (s *service) generateItemID(ctx context.Context, collectionID types.U64) (types.U128, error) {
	var itemID types.U128

	for {
		select {
		case <-ctx.Done():
			return itemID, ctx.Err()
		default:
			itemID = types.NewU128(*big.NewInt(int64(rand.Int())))

			_, err := s.api.GetItemDetails(collectionID, itemID)

			if err != nil {
				if errors.Is(err, uniques.ErrItemDetailsNotFound) {
					return itemID, nil
				}

				return itemID, err
			}
		}
	}
}

func getNFTMintRunnerJob(documentPending bool, args []any) *gocelery.Job {
	description := "Mint NFT on Centrifuge Chain"
	runner := mintNFTV3Job
	task := "add_nft_v3_to_document"

	if documentPending {
		description = "Commit pending document and mint NFT on Centrifuge Chain"
		runner = commitAndMintNFTV3Job
		task = "commit_pending_document"
	}

	return gocelery.NewRunnerJob(
		description,
		runner,
		task,
		args,
		make(map[string]any),
		time.Time{},
	)
}
