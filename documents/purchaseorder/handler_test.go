// +build unit

package purchaseorder

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	Service
	mock.Mock
}

func (m mockService) Create(ctx *header.ContextHeader, doc documents.Model) (documents.Model, error) {
	args := m.Called(ctx, doc)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m mockService) Update(ctx *header.ContextHeader, doc documents.Model) (documents.Model, error) {
	args := m.Called(ctx, doc)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m mockService) DeriveFromCreatePayload(req *clientpopb.PurchaseOrderCreatePayload, ctxh *header.ContextHeader) (documents.Model, error) {
	args := m.Called(req, ctxh)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m mockService) GetCurrentVersion(identifier []byte) (documents.Model, error) {
	args := m.Called(identifier)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m mockService) GetVersion(identifier []byte, version []byte) (documents.Model, error) {
	args := m.Called(identifier, version)
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

func (m mockService) DeriveFromUpdatePayload(payload *clientpopb.PurchaseOrderUpdatePayload, ctxh *header.ContextHeader) (documents.Model, error) {
	args := m.Called(payload, ctxh)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func TestGRPCHandler_Create(t *testing.T) {
	h := getHandler()
	req := testingdocuments.CreatePOPayload()
	ctx := context.Background()
	model := &testingdocuments.MockModel{}
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)

	// derive fails
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(nil, fmt.Errorf("derive failed")).Once()
	h.service = srv
	resp, err := h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive failed")

	// create fails
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Create", ctxh, model).Return(nil, fmt.Errorf("create failed")).Once()
	h.service = srv
	resp, err = h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")

	// derive response fails
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Create", ctxh, model).Return(model, nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(nil, fmt.Errorf("derive response fails")).Once()
	h.service = srv
	resp, err = h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive response fails")

	// success
	eresp := &clientpopb.PurchaseOrderResponse{}
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Create", ctxh, model).Return(model, nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(eresp, nil).Once()
	h.service = srv
	resp, err = h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, eresp, resp)
}

func TestGrpcHandler_Update(t *testing.T) {
	h := getHandler()
	p := testingdocuments.CreatePOPayload()
	req := &clientpopb.PurchaseOrderUpdatePayload{
		Data:          p.Data,
		Collaborators: p.Collaborators,
	}
	ctx := context.Background()
	model := &testingdocuments.MockModel{}
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)

	// derive fails
	srv := h.service.(*mockService)
	srv.On("DeriveFromUpdatePayload", req, ctxh).Return(nil, fmt.Errorf("derive failed")).Once()
	h.service = srv
	resp, err := h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive failed")

	// create fails
	srv.On("DeriveFromUpdatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Update", ctxh, model).Return(nil, fmt.Errorf("update failed")).Once()
	h.service = srv
	resp, err = h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")

	// derive response fails
	srv.On("DeriveFromUpdatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Update", ctxh, model).Return(model, nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(nil, fmt.Errorf("derive response fails")).Once()
	h.service = srv
	resp, err = h.Update(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive response fails")

	// success
	eresp := &clientpopb.PurchaseOrderResponse{}
	srv.On("DeriveFromUpdatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Update", ctxh, model).Return(model, nil).Once()
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
	return &grpcHandler{service: &mockService{}, config: cfg}
}

func TestGrpcHandler_Get(t *testing.T) {
	identifier := "0x01010101"
	identifierBytes, _ := hexutil.Decode(identifier)
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientpopb.GetRequest{Identifier: identifier}
	response := &clientpopb.PurchaseOrderResponse{}
	srv.On("GetCurrentVersion", identifierBytes).Return(model, nil)
	srv.On("DerivePurchaseOrderResponse", model).Return(response, nil)
	res, err := h.Get(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}

func TestGrpcHandler_GetVersion_invalid_input(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	payload := &clientpopb.GetVersionRequest{Identifier: "0x0x", Version: "0x00"}
	res, err := h.GetVersion(context.Background(), payload)
	assert.EqualError(t, err, "identifier is invalid: invalid hex string")
	payload.Version = "0x0x"
	payload.Identifier = "0x01"

	res, err = h.GetVersion(context.Background(), payload)
	assert.EqualError(t, err, "version is invalid: invalid hex string")
	payload.Version = "0x00"
	payload.Identifier = "0x01"

	mockErr := fmt.Errorf("not found")
	srv.On("GetVersion", []byte{0x01}, []byte{0x00}).Return(nil, mockErr)
	res, err = h.GetVersion(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.EqualError(t, err, "document not found: not found")
	assert.Nil(t, res)
}

func TestGrpcHandler_GetVersion(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientpopb.GetVersionRequest{Identifier: "0x01", Version: "0x00"}

	response := &clientpopb.PurchaseOrderResponse{}
	srv.On("GetVersion", []byte{0x01}, []byte{0x00}).Return(model, nil)
	srv.On("DerivePurchaseOrderResponse", model).Return(response, nil)
	res, err := h.GetVersion(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, response)
}
