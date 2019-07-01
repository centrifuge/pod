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

// Service defines the functionality for the CoreAPI service.
type Service struct {
	DocSrv      documents.Service
	JobsSrv     jobs.Manager
	NFTSrv      nft.Service
	AccountsSrv config.Service
}

// CreateDocument creates the document from the payload and anchors it.
func (s Service) CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	return s.DocSrv.CreateModel(ctx, payload)
}

// UpdateDocument updates the document from the payload and anchors the next version.
func (s Service) UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	return s.DocSrv.UpdateModel(ctx, payload)
}

// GetJobStatus returns the job status.
func (s Service) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	return s.JobsSrv.GetJobStatus(account, id)
}

// GetDocument returns the latest version of the document.
func (s Service) GetDocument(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.DocSrv.GetCurrentVersion(ctx, docID)
}

// GetDocumentVersion returns the specific version of the document
func (s Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	return s.DocSrv.GetVersion(ctx, docID, versionID)
}

// GenerateProofs returns the proofs for the latest version of the document.
func (s Service) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.DocSrv.CreateProofs(ctx, docID, fields)
}

// GenerateProofsForVersion returns the proofs for the specific version of the document.
func (s Service) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.DocSrv.CreateProofsForVersion(ctx, docID, versionID, fields)
}

// MintNFT mints an NFT.
func (s Service) MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error) {
	resp, _, err := s.NFTSrv.MintNFT(ctx, request)
	return resp, err
}

// TransferNFT transfers NFT with tokenID in a given registry to `to` address.
func (s Service) TransferNFT(ctx context.Context, to, registry common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error) {
	resp, _, err := s.NFTSrv.TransferFrom(ctx, registry, to, tokenID)
	return resp, err
}

// OwnerOfNFT returns the owner of the NFT.
func (s Service) OwnerOfNFT(registry common.Address, tokenID nft.TokenID) (common.Address, error) {
	return s.NFTSrv.OwnerOf(registry, tokenID[:])
}

// SignPayload uses the accountID's secret key to sign the payload and returns the signature
func (s Service) SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	return s.AccountsSrv.Sign(accountID, payload)
}
