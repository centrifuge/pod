// +build unit

package purchaseorder

import (
	"context"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	Service
	mock.Mock
}

func (m mockService) Create(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.Job(ctx), nil, args.Error(2)
}

func (m mockService) Update(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.Job(ctx), nil, args.Error(2)
}

func (m mockService) DeriveFromCreatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m mockService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	args := m.Called(ctx, documentID)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m mockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	args := m.Called(ctx, documentID, version)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m mockService) DerivePurchaseOrderData(po documents.Model) (*clientpopb.PurchaseOrderData, error) {
	args := m.Called(po)
	data, _ := args.Get(0).(*clientpopb.PurchaseOrderData)
	return data, args.Error(1)
}

func (m mockService) DerivePurchaseOrderResponse(po documents.Model) (*clientpopb.PurchaseOrderResponse, error) {
	args := m.Called(po)
	data, _ := args.Get(0).(*clientpopb.PurchaseOrderResponse)
	return data, args.Error(1)
}

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func getHandler() *grpcHandler {
	return &grpcHandler{service: &mockService{}, config: configService}
}

func TestGrpcHandler_GetVersion_invalid_input(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	payload := &clientpopb.GetVersionRequest{DocumentId: "0x0x", VersionId: "0x00"}
	res, err := h.GetVersion(testingconfig.HandlerContext(configService), payload)
	assert.Error(t, err)
	assert.EqualError(t, err, "identifier is invalid: invalid hex string")
	payload.VersionId = "0x0x"
	payload.DocumentId = "0x01"

	res, err = h.GetVersion(testingconfig.HandlerContext(configService), payload)
	assert.Error(t, err)
	assert.EqualError(t, err, "version is invalid: invalid hex string")
	payload.VersionId = "0x00"
	payload.DocumentId = "0x01"

	mockErr := errors.New("not found")
	srv.On("GetVersion", mock.Anything, []byte{0x01}, []byte{0x00}).Return(nil, mockErr)
	res, err = h.GetVersion(testingconfig.HandlerContext(configService), payload)
	srv.AssertExpectations(t)
	assert.EqualError(t, err, "document not found: not found")
	assert.Nil(t, res)
}

func TestGrpcHandler_GetVersion(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientpopb.GetVersionRequest{DocumentId: "0x01", VersionId: "0x00"}

	response := &clientpopb.PurchaseOrderResponse{}
	srv.On("GetVersion", mock.Anything, []byte{0x01}, []byte{0x00}).Return(model, nil)
	srv.On("DerivePurchaseOrderResponse", model).Return(response, nil)
	res, err := h.GetVersion(testingconfig.HandlerContext(configService), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, response)
}
