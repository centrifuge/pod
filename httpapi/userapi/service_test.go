// +build unit

package userapi

import (
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/stretchr/testify/mock"
)

type mockTransferService struct {
	transferdetails.Service
	mock.Mock
}

//func TestService_CreateDocument(t *testing.T) {
//	transferSrv := new(mockTransferService)
//	service := Service{docSrv: nil, transferDetailsService: transferSrv}
//	m := new(testingdocuments.MockModel)
//	transferSrv.On("CreateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
//	nm, _, err := service.CreateTransferDetail(context.Background(), transferdetails.CreateTransferDetailRequest{})
//	assert.NoError(t, err)
//	assert.Equal(t, m, nm)
//}

//func TestService_UpdateDocument(t *testing.T) {
//	docSrv := new(testingdocuments.MockService)
//	srv := Service{docSrv: docSrv}
//	m := new(testingdocuments.MockModel)
//	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
//	nm, _, err := srv.UpdateDocument(context.Background(), documents.UpdatePayload{})
//	assert.NoError(t, err)
//	assert.Equal(t, m, nm)
//}
