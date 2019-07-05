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
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
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

func (m mockService) DeriveFromUpdatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func TestGrpcHandler_Update(t *testing.T) {
	h := getHandler()
	p := testingdocuments.CreatePOPayload()
	req := &clientpopb.PurchaseOrderUpdatePayload{
		Data:        p.Data,
		ReadAccess:  p.ReadAccess,
		WriteAccess: p.WriteAccess,
	}
	ctx := testingconfig.HandlerContext(configService)
	model := &testingdocuments.MockModel{}

	// derive fails
	srv := h.service.(*mockService)
	srv.On("DeriveFromUpdatePayload", mock.Anything, req).Return(nil, errors.New("derive failed")).Once()
	h.service = srv
	resp, err := h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive failed")

	// create fails
	srv.On("DeriveFromUpdatePayload", mock.Anything, req).Return(model, nil).Once()
	srv.On("Update", mock.Anything, model).Return(nil, jobs.NilJobID().String(), errors.New("update failed")).Once()
	h.service = srv
	resp, err = h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")

	// derive response fails
	srv.On("DeriveFromUpdatePayload", mock.Anything, req).Return(model, nil).Once()
	srv.On("Update", mock.Anything, model).Return(model, jobs.NilJobID().String(), nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(nil, errors.New("derive response fails")).Once()
	h.service = srv
	resp, err = h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive response fails")

	// success
	eresp := &clientpopb.PurchaseOrderResponse{Header: new(documentpb.ResponseHeader)}
	srv.On("DeriveFromUpdatePayload", mock.Anything, req).Return(model, nil).Once()
	srv.On("Update", mock.Anything, model).Return(model, jobs.NilJobID().String(), nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(eresp, nil).Once()
	h.service = srv
	resp, err = h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, eresp, resp)
}

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func getHandler() *grpcHandler {
	return &grpcHandler{service: &mockService{}, config: configService}
}
