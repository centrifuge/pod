package coreapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
)

// NewService returns the new CoreAPI Service.
func NewService(docSrv documents.Service, jobsSrv jobs.Manager, nftSrv nft.Service, accountsSrv config.Service) Service {
	return Service{
		docSrv:      docSrv,
		jobsSrv:     jobsSrv,
		nftSrv:      nftSrv,
		accountsSrv: accountsSrv,
	}
}

// Service defines the functionality for the CoreAPI service.
type Service struct {
	docSrv      documents.Service
	jobsSrv     jobs.Manager
	nftSrv      nft.Service
	accountsSrv config.Service
}

// CreateDocument creates the document from the payload and anchors it.
func (s Service) CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Document, jobs.JobID, error) {
	return s.docSrv.CreateModel(ctx, payload)
}

// UpdateDocument updates the document from the payload and anchors the next version.
func (s Service) UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Document, jobs.JobID, error) {
	return s.docSrv.UpdateModel(ctx, payload)
}

// GetJobStatus returns the job status.
func (s Service) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	return s.jobsSrv.GetJobStatus(account, id)
}

// GetDocument returns the latest version of the document.
func (s Service) GetDocument(ctx context.Context, docID []byte) (documents.Document, error) {
	return s.docSrv.GetCurrentVersion(ctx, docID)
}

// GetDocumentVersion returns the specific version of the document
func (s Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Document, error) {
	return s.docSrv.GetVersion(ctx, docID, versionID)
}

// GenerateProofs returns the proofs for the latest version of the document.
func (s Service) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docSrv.CreateProofs(ctx, docID, fields)
}

// GenerateProofsForVersion returns the proofs for the specific version of the document.
func (s Service) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docSrv.CreateProofsForVersion(ctx, docID, versionID, fields)
}

// GetAccount returns the Account associated with accountID
func (s Service) GetAccount(accountID []byte) (config.Account, error) {
	return s.accountsSrv.GetAccount(accountID)
}

// GetAccounts returns all the accounts.
func (s Service) GetAccounts() ([]config.Account, error) {
	return s.accountsSrv.GetAccounts()
}

// CreateAccount creates a new account from the data provided.
func (s Service) CreateAccount(acc config.Account) (config.Account, error) {
	return s.accountsSrv.CreateAccount(acc)
}

// UpdateAccount updates the existing account with the data provided.
func (s Service) UpdateAccount(acc config.Account) (config.Account, error) {
	return s.accountsSrv.UpdateAccount(acc)
}
