// +build unit

package funding

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var configService config.Service

func TestGRPCHandler_Update(t *testing.T) {
	srv := &MockService{}

	h := &grpcHandler{service: srv, config: configService}
	jobID := jobs.NewJobID()

	// successful
	srv.On("DeriveFromUpdatePayload", mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("Update", mock.Anything, mock.Anything).Return(nil, jobID, nil).Once()
	srv.On("DeriveFundingResponse", mock.Anything, mock.Anything, mock.Anything).Return(&clientfunpb.FundingResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()

	response, err := h.Update(testingconfig.HandlerContext(configService), &clientfunpb.FundingUpdatePayload{DocumentId: hexutil.Encode(utils.RandomSlice(32)), Data: &clientfunpb.FundingData{Currency: "eur"}})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestGRPCHandler_Sign(t *testing.T) {
	srv := &MockService{}
	h := &grpcHandler{service: srv, config: configService}
	jobID := jobs.NewJobID()

	// successful
	srv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("DeriveFundingResponse", mock.Anything, mock.Anything, mock.Anything).Return(&clientfunpb.FundingResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()
	srv.On("Sign", mock.Anything, mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil).Once()
	srv.On("Update", mock.Anything, mock.Anything).Return(nil, jobID, nil).Once()

	response, err := h.Sign(testingconfig.HandlerContext(configService), &clientfunpb.Request{DocumentId: hexutil.Encode(utils.RandomSlice(32)), AgreementId: hexutil.Encode(utils.RandomSlice(32))})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// fail
	response, err = h.Sign(testingconfig.HandlerContext(configService), &clientfunpb.Request{AgreementId: hexutil.Encode(utils.RandomSlice(32))})
	assert.Error(t, err)
}

func TestGRPCHandler_Get(t *testing.T) {
	srv := &MockService{}
	h := &grpcHandler{service: srv, config: configService}

	srv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("DeriveFundingResponse", mock.Anything, mock.Anything, mock.Anything).Return(&clientfunpb.FundingResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()

	response, err := h.Get(testingconfig.HandlerContext(configService), &clientfunpb.Request{DocumentId: hexutil.Encode(utils.RandomSlice(32)), AgreementId: hexutil.Encode(utils.RandomSlice(32))})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestGRPCHandler_GetVersion(t *testing.T) {
	srv := &MockService{}
	h := &grpcHandler{service: srv, config: configService}

	srv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("DeriveFundingResponse", mock.Anything, mock.Anything, mock.Anything).Return(&clientfunpb.FundingResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()

	response, err := h.GetVersion(testingconfig.HandlerContext(configService), &clientfunpb.GetVersionRequest{DocumentId: hexutil.Encode(utils.RandomSlice(32)), VersionId: hexutil.Encode(utils.RandomSlice(32)), AgreementId: hexutil.Encode(utils.RandomSlice(32))})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestGRPCHandler_GetList(t *testing.T) {
	srv := &MockService{}
	h := &grpcHandler{service: srv, config: configService}

	srv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("DeriveFundingListResponse", mock.Anything, mock.Anything).Return(&clientfunpb.FundingListResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()

	response, err := h.GetListVersion(testingconfig.HandlerContext(configService), &clientfunpb.GetListVersionRequest{DocumentId: hexutil.Encode(utils.RandomSlice(32)), VersionId: hexutil.Encode(utils.RandomSlice(32))})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}
