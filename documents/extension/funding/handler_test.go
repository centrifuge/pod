// +build unit

package funding

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	Service
	mock.Mock
}

var configService config.Service

func (m *mockService) DeriveFromPayload(ctx context.Context, payload *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *mockService) DeriveFundingResponse(doc documents.Model, payload *clientfundingpb.FundingCreatePayload) (*clientfundingpb.FundingResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientfundingpb.FundingResponse)
	return data, args.Error(1)
}

func (m *mockService) Update(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.Job(ctx), nil, args.Error(2)
}

func TestGRPCHandler_Create(t *testing.T) {
	srv := &mockService{}

	h := &grpcHandler{service: srv, config: configService}
	jobID := jobs.NewJobID()

	// no identifier
	response, err := h.Create(testingconfig.HandlerContext(configService), &clientfundingpb.FundingCreatePayload{Data: &clientfundingpb.FundingData{Currency: "eur"}})
	assert.Nil(t, response)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "empty hex string")

	// successful
	srv.On("DeriveFromPayload", mock.Anything, mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("Update", mock.Anything, mock.Anything).Return(nil, jobID, nil).Once()
	srv.On("DeriveFundingResponse", mock.Anything, mock.Anything).Return(&clientfundingpb.FundingResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()

	response, err = h.Create(testingconfig.HandlerContext(configService), &clientfundingpb.FundingCreatePayload{Identifier: hexutil.Encode(utils.RandomSlice(32)), Data: &clientfundingpb.FundingData{Currency: "eur"}})
	assert.NoError(t, err)
	assert.NotNil(t, response)

}
