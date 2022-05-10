package v3

import (
	"context"
	"encoding/gob"
	"errors"
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
)

func init() {
	gob.Register(types.U64(0))
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
	DocumentID     []byte
	ClassID        types.U64
	Owner          types.AccountID // substrate account ID
	Metadata       string
	FreezeMetadata bool
}

// MintNFTResponse is the response object for a MintNFTRequest, it holds the job ID and instance ID of the NFT.
type MintNFTResponse struct {
	JobID      string
	InstanceID types.U128
}

// CreateNFTClassRequest is the response object for creating an NFT class on Centrifuge chain.
type CreateNFTClassRequest struct {
	// TODO(cdamian): Add more fields such as admin?
	ClassID types.U64
}

// CreateNFTClassResponse is the response object for a CreateNFTClassRequest, it holds the job ID and the newly created class ID.
type CreateNFTClassResponse struct {
	JobID   string
	ClassID types.U64
}

type InstanceMetadataOf struct {
	ClassID    types.U64
	InstanceID types.U128
}

type Service interface {
	CreateNFTClass(ctx context.Context, req *CreateNFTClassRequest) (*CreateNFTClassResponse, error)
	MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error)
	OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error)
	InstanceMetadataOf(ctx context.Context, req *InstanceMetadataOf) (*types.InstanceMetadata, error)
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

		return nil, ErrInstanceIDGeneration
	}

	did, err := identity.NewDIDFromBytes(acc.GetIdentityID())

	if err != nil {
		s.log.Errorf("Couldn't generate identity: %s", err)

		return nil, ErrIdentityRetrieval
	}

	jobID, err := s.dispatchNFTMintJob(did, instanceID, req)

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

			return ErrClassIDDecoding
		}

		if nftClassID != req.ClassID {
			continue
		}

		var instanceID types.U128

		if err := types.DecodeFromBytes(nft.InstanceId, &instanceID); err != nil {
			s.log.Errorf("Couldn't decode instance ID: %s", err)

			return ErrInstanceIDDecoding
		}

		_, err := s.api.GetInstanceDetails(ctx, nftClassID, instanceID)

		if err != nil {
			if errors.Is(err, ErrInstanceDetailsNotFound) {
				s.log.Info("NFT instance found but not minted")

				return nil
			}

			s.log.Errorf("Couldn't get instance details: %s", err)

			return ErrInstanceDetailsRetrieval
		}

		docVersion := doc.CurrentVersion()

		anchorID, err := anchors.ToAnchorID(docVersion)

		if err != nil {
			s.log.Errorf("Couldn't parse anchor ID for doc with version %s: %s", docVersion, err)
		}

		err = ErrInstanceAlreadyMinted

		return fmt.Errorf("instance ID %d was already minted for doc with anchor %s: %w", instanceID, anchorID, err)

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

	if err := s.dispatchJob(did, job); err != nil {
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

			_, err := s.api.GetInstanceDetails(ctx, classID, instanceID)

			if err != nil {
				if errors.Is(err, ErrInstanceDetailsNotFound) {
					return instanceID, nil
				}

				return instanceID, err
			}
		}
	}
}

func (s *service) OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error) {
	instanceDetails, err := s.api.GetInstanceDetails(ctx, req.ClassID, req.InstanceID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve the instance details: %s", err)

		if errors.Is(err, ErrInstanceDetailsNotFound) {
			return nil, ErrOwnerNotFound
		}

		return nil, ErrInstanceDetailsRetrieval
	}

	return &OwnerOfResponse{
		ClassID:    req.ClassID,
		InstanceID: req.InstanceID,
		AccountID:  instanceDetails.Owner,
	}, nil
}

func (s *service) CreateNFTClass(ctx context.Context, req *CreateNFTClassRequest) (*CreateNFTClassResponse, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		s.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrAccountFromContextRetrieval
	}

	did, err := identity.NewDIDFromBytes(acc.GetIdentityID())

	if err != nil {
		s.log.Errorf("Couldn't generate identity: %s", err)

		return nil, ErrIdentityRetrieval
	}

	classExists, err := s.classExists(ctx, req.ClassID)

	if err != nil {
		s.log.Errorf("Couldn't check if class already exists: %s", err)

		return nil, ErrClassCheck
	}

	if classExists {
		s.log.Errorf("Class already exists")

		return nil, ErrClassAlreadyExists
	}

	jobID, err := s.dispatchCreateClassJob(did, req.ClassID)

	if err != nil {
		s.log.Errorf("Couldn't create class: %s", err)

		return nil, ErrCreateClassJobDispatch
	}

	return &CreateNFTClassResponse{
		JobID:   jobID.Hex(),
		ClassID: req.ClassID,
	}, nil
}

func (s *service) classExists(ctx context.Context, classID types.U64) (bool, error) {
	_, err := s.api.GetClassDetails(ctx, classID)

	if err != nil {
		if errors.Is(err, ErrClassDetailsNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *service) dispatchCreateClassJob(did identity.DID, classID types.U64) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Create NFT class on Centrifuge Chain",
		createNFTClassV3Job,
		"create_nft_class_v3",
		[]interface{}{
			did,
			classID,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if err := s.dispatchJob(did, job); err != nil {
		s.log.Errorf("Couldn't dispatch create class job: %s", err)

		return nil, fmt.Errorf("failed to dispatch create class job: %w", err)
	}

	return job.ID, nil
}

func (s *service) dispatchJob(did identity.DID, job *gocelery.Job) error {
	if _, err := s.dispatcher.Dispatch(did, job); err != nil {
		return err
	}

	return nil
}

func (s *service) InstanceMetadataOf(ctx context.Context, req *InstanceMetadataOf) (*types.InstanceMetadata, error) {
	instanceMetadata, err := s.api.GetInstanceMetadata(ctx, req.ClassID, req.InstanceID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve instance metadata: %s", err)

		if errors.Is(err, ErrInstanceMetadataNotFound) {
			return nil, err
		}

		return nil, ErrInstanceMetadataRetrieval
	}

	return instanceMetadata, nil
}
