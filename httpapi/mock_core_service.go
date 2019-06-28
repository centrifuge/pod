// +build unit

package httpapi

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockCoreService struct {
	mock.Mock
}

func (m MockCoreService) CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model := args.Get(0).(documents.Model)
	job := args.Get(1).(jobs.JobID)
	return model, job, args.Error(2)
}

func (m MockCoreService) UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model := args.Get(0).(documents.Model)
	job := args.Get(1).(jobs.JobID)
	return model, job, args.Error(2)
}

func (m MockCoreService) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	args := m.Called(account, id)
	job := args.Get(0).(jobs.StatusResponse)
	return job, args.Error(1)
}

func (m MockCoreService) GetDocument(ctx context.Context, docID []byte) (documents.Model, error) {
	args := m.Called(ctx, docID)
	model := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m MockCoreService) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	args := m.Called(ctx, docID, versionID)
	model := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m MockCoreService) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(ctx, docID, fields)
	model := args.Get(0).(*documents.DocumentProof)
	return model, args.Error(1)
}

func (m MockCoreService) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(ctx, docID, fields)
	model := args.Get(0).(*documents.DocumentProof)
	return model, args.Error(1)
}

func (m MockCoreService) MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error) {
	args := m.Called(ctx, request)
	model := args.Get(0).(*nft.TokenResponse)
	return model, args.Error(1)
}

func (m MockCoreService) TransferNFT(ctx context.Context, to, registry common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error) {
	args := m.Called(ctx, to, registry, tokenID)
	model := args.Get(0).(*nft.TokenResponse)
	return model, args.Error(1)
}

func (m MockCoreService) OwnerOfNFT(registry common.Address, tokenID nft.TokenID) (common.Address, error) {
	args := m.Called(registry, tokenID)
	model := args.Get(0).(common.Address)
	return model, args.Error(1)
}

func (m MockCoreService) SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	args := m.Called(accountID, payload)
	model := args.Get(0).(*coredocumentpb.Signature)
	return model, args.Error(1)
}
