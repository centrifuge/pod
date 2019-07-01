package coreapi

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
)

// Service provides functionality for Core APIs.
type Service interface {
	// CreateDocument creates the document from the payload and anchors it.
	CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error)

	// UpdateDocument updates the document from the payload and anchors the next version.
	UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error)

	// GetJobStatus returns the job status.
	GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error)

	// GetDocument returns the latest version of the document.
	GetDocument(ctx context.Context, docID []byte) (documents.Model, error)

	// GetDocumentVersion returns the specific version of the document
	GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error)

	// GenerateProofs returns the proofs for the latest version of the document.
	GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error)

	// GenerateProofsForVersion returns the proofs for the specific version of the document.
	GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error)

	// MintNFT mints an NFT.
	MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error)

	// TransferNFT transfers NFT with tokenID in a given registry to `to` address.
	TransferNFT(ctx context.Context, to, registry common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error)

	// OwnerOfNFT returns the owner of the NFT.
	OwnerOfNFT(registry common.Address, tokenID nft.TokenID) (common.Address, error)

	// SignPayload uses the accountID's secret key to sign the payload and returns the signature
	SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error)
}

// DefaultService returns the default implementation of the service
func DefaultService(docSrv documents.Service, manager jobs.Manager, nftSrv nft.Service, accountSrv config.Service) Service {
	return service{
		docService:      docSrv,
		jobsService:     manager,
		nftService:      nftSrv,
		accountsService: accountSrv,
	}
}

type service struct {
	docService      documents.Service
	jobsService     jobs.Manager
	nftService      nft.Service
	accountsService config.Service
}

// CreateDocument creates the document from the payload and anchors it.
func (s service) CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	return s.docService.CreateModel(ctx, payload)
}

// UpdateDocument updates the document from the payload and anchors the next version.
func (s service) UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	return s.docService.UpdateModel(ctx, payload)
}

// GetJobStatus returns the job status.
func (s service) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	return s.jobsService.GetJobStatus(account, id)
}

// GetDocument returns the latest version of the document.
func (s service) GetDocument(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.docService.GetCurrentVersion(ctx, docID)
}

// GetDocumentVersion returns the specific version of the document
func (s service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	return s.docService.GetVersion(ctx, docID, versionID)
}

// GenerateProofs returns the proofs for the latest version of the document.
func (s service) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docService.CreateProofs(ctx, docID, fields)
}

// GenerateProofsForVersion returns the proofs for the specific version of the document.
func (s service) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docService.CreateProofsForVersion(ctx, docID, versionID, fields)
}

// MintNFT mints an NFT.
func (s service) MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error) {
	resp, _, err := s.nftService.MintNFT(ctx, request)
	return resp, err
}

// TransferNFT transfers NFT with tokenID in a given registry to `to` address.
func (s service) TransferNFT(ctx context.Context, to, registry common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error) {
	resp, _, err := s.nftService.TransferFrom(ctx, registry, to, tokenID)
	return resp, err
}

// OwnerOfNFT returns the owner of the NFT.
func (s service) OwnerOfNFT(registry common.Address, tokenID nft.TokenID) (common.Address, error) {
	return s.nftService.OwnerOf(registry, tokenID[:])
}

// SignPayload uses the accountID's secret key to sign the payload and returns the signature
func (s service) SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	return s.accountsService.Sign(accountID, payload)
}
