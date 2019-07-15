// +build unit integration testworld

package funding

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

type MockService struct {
	Service
	mock.Mock
}

func (m *MockService) DeriveFromPayload(ctx context.Context, req *clientfunpb.FundingCreatePayload) (documents.Model, error) {
	args := m.Called(ctx, req)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) DeriveFromUpdatePayload(ctx context.Context, req *clientfunpb.FundingUpdatePayload) (documents.Model, error) {
	args := m.Called(ctx, req)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) DeriveFundingResponse(ctx context.Context, doc documents.Model, fundingId string) (*clientfunpb.FundingResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientfunpb.FundingResponse)
	return data, args.Error(1)
}

func (m *MockService) DeriveFundingListResponse(ctx context.Context, doc documents.Model) (*clientfunpb.FundingListResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientfunpb.FundingListResponse)
	return data, args.Error(1)
}

func (m *MockService) Update(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *MockService) GetCurrentVersion(ctx context.Context, identifier []byte) (documents.Model, error) {
	args := m.Called(ctx, identifier)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) GetVersion(ctx context.Context, identifier, version []byte) (documents.Model, error) {
	args := m.Called(ctx, identifier)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error) {
	args := m.Called(ctx, fundingID, identifier)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) CreateFundingAgreement(ctx context.Context, docID []byte, data *Data) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, docID, data)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockService) GetDataAndSignatures(ctx context.Context, model documents.Model, fundingID string) (Data, []Signature, error) {
	args := m.Called(ctx, model, fundingID)
	d, _ := args.Get(0).(Data)
	sigs, _ := args.Get(1).([]Signature)
	return d, sigs, args.Error(2)
}
