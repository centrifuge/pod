// +build unit

package userapi

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransferService struct {
	transferdetails.Service
	mock.Mock
}

func (m *MockTransferService) CreateTransferDetail(ctx context.Context, payload transferdetails.CreateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockTransferService) UpdateTransferDetail(ctx context.Context, payload transferdetails.UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockTransferService) DeriveTransferDetail(ctx context.Context, model documents.Model, transferID []byte) (*transferdetails.TransferDetail, documents.Model, error) {
	args := m.Called(ctx, model, transferID)
	td, _ := args.Get(0).(*transferdetails.TransferDetail)
	nm, _ := args.Get(1).(documents.Model)
	return td, nm, args.Error(2)
}

func (m *MockTransferService) DeriveTransferList(ctx context.Context, model documents.Model) (*transferdetails.TransferDetailList, documents.Model, error) {
	args := m.Called(ctx, model)
	td, _ := args.Get(0).(*transferdetails.TransferDetailList)
	nm, _ := args.Get(1).(documents.Model)
	return td, nm, args.Error(2)
}

func TestService_CreateTransferDetail(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	transferSrv.On("CreateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	nm, _, err := service.CreateTransferDetail(context.Background(), transferdetails.CreateTransferDetailRequest{
		DocumentID: "test_id",
		Data:       &transferdetails.TransferDetailData{},
	})
	assert.NoError(t, err)
	assert.Equal(t, m, nm)
}

func TestService_UpdateDocument(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	transferSrv.On("UpdateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	nm, _, err := service.UpdateTransferDetail(context.Background(), transferdetails.UpdateTransferDetailRequest{
		DocumentID: "test_id",
		TransferID: "test_transfer",
		Data:       &transferdetails.TransferDetailData{},
	})
	assert.NoError(t, err)
	assert.Equal(t, m, nm)
}

func TestService_GetCurrentTransferDetail(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetail)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferDetail", mock.Anything, mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetCurrentTransferDetail(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_GetCurrentTransferDetailList(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetailList)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferList", mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetCurrentTransferDetailsList(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_GetVersionTransferDetail(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetail)
	docSrv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferDetail", mock.Anything, mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetCurrentTransferDetail(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_GetVersionTransferDetailList(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetailList)
	docSrv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferList", mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetVersionTransferDetailsList(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}
