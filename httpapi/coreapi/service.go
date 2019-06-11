package coreapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
)

// Service provides functionality for Core APIs.
type Service struct {
	docService  documents.Service
	jobsService jobs.Manager
	nftService  nft.Service
}

// CreateDocument creates the document from the payload and anchors it.
func (s Service) CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	return s.docService.CreateModel(ctx, payload)
}

// UpdateDocument updates the document from the payload and anchors the next version.
func (s Service) UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	return s.docService.UpdateModel(ctx, payload)
}

// GetJobStatus returns the job status.
func (s Service) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	return s.jobsService.GetJobStatus(account, id)
}

// GetDocument returns the latest version of the document.
func (s Service) GetDocument(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.docService.GetCurrentVersion(ctx, docID)
}

// GetDocumentVersion returns the specific version of the document
func (s Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	return s.docService.GetVersion(ctx, docID, versionID)
}

// GenerateProofs returns the proofs for the latest version of the document.
func (s Service) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docService.CreateProofs(ctx, docID, fields)
}

// GenerateProofsForVersion returns the proofs for the specific version of the document.
func (s Service) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docService.CreateProofsForVersion(ctx, docID, versionID, fields)
}

// MintNFT mints an NFT.
func (s Service) MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error) {
	resp, _, err := s.nftService.MintNFT(ctx, request)
	return resp, err
}

// TransferNFT transfers NFT with tokenID in a given registry to `to` address.
func (s Service) TransferNFT(ctx context.Context, to, registry common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error) {
	resp, _, err := s.nftService.TransferFrom(ctx, registry, to, tokenID)
	return resp, err
}
