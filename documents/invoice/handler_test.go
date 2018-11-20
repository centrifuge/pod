// +build unit

package invoice

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	Service
	mock.Mock
}

func (m *mockService) DeriveFromCreatePayload(payload *clientinvoicepb.InvoiceCreatePayload, contextHeader *header.ContextHeader) (documents.Model, error) {
	args := m.Called(payload, contextHeader)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *mockService) Create(ctx *header.ContextHeader, inv documents.Model) (documents.Model, error) {
	args := m.Called(ctx, inv)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *mockService) GetCurrentVersion(identifier []byte) (documents.Model, error) {
	args := m.Called(identifier)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *mockService) GetVersion(identifier []byte, version []byte) (documents.Model, error) {
	args := m.Called(identifier, version)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *mockService) DeriveInvoiceData(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientinvoicepb.InvoiceData)
	return data, args.Error(1)
}

func (m *mockService) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientinvoicepb.InvoiceResponse)
	return data, args.Error(1)
}

func (m *mockService) Update(ctx *header.ContextHeader, doc documents.Model) (documents.Model, error) {
	args := m.Called(ctx, doc)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, args.Error(1)
}

func (m *mockService) DeriveFromUpdatePayload(payload *clientinvoicepb.InvoiceUpdatePayload, contextHeader *header.ContextHeader) (documents.Model, error) {
	args := m.Called(payload, contextHeader)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func getHandler() *grpcHandler {
	return &grpcHandler{service: &mockService{}, config: cfg}
}

func TestGRPCHandler_Create_derive_fail(t *testing.T) {
	// DeriveFrom payload fails
	h := getHandler()
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("derive failed")).Once()
	_, err := h.Create(context.Background(), nil)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "derive failed")
}

func TestGRPCHandler_Create_create_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(new(Invoice), nil).Once()
	srv.On("Create", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("create failed")).Once()
	payload := &clientinvoicepb.InvoiceCreatePayload{Data: &clientinvoicepb.InvoiceData{GrossAmount: 300}}
	_, err := h.Create(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "create failed")
}

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func (m *mockModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(*coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func (m *mockModel) JSON() ([]byte, error) {
	args := m.Called()
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

func TestGRPCHandler_Create_DeriveInvoiceResponse_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(Invoice)
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("Create", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("DeriveInvoiceResponse", mock.Anything).Return(nil, fmt.Errorf("derive response failed"))
	payload := &clientinvoicepb.InvoiceCreatePayload{Data: &clientinvoicepb.InvoiceData{Currency: "EUR"}}
	_, err := h.Create(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "derive response failed")
}

func TestGrpcHandler_Create(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(Invoice)
	payload := &clientinvoicepb.InvoiceCreatePayload{Data: &clientinvoicepb.InvoiceData{GrossAmount: 300}, Collaborators: []string{"0x010203040506"}}
	response := &clientinvoicepb.InvoiceResponse{}
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("Create", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("DeriveInvoiceResponse", model).Return(response, nil)
	res, err := h.Create(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}

func TestGrpcHandler_Get_invalid_input(t *testing.T) {
	identifier := "0x01010101"
	identifierBytes, _ := hexutil.Decode(identifier)
	h := getHandler()
	srv := h.service.(*mockService)
	payload := &clientinvoicepb.GetRequest{Identifier: "invalid"}

	res, err := h.Get(context.Background(), payload)
	assert.Nil(t, res)
	assert.EqualError(t, err, "identifier is an invalid hex string: hex string without 0x prefix")

	payload.Identifier = identifier
	srv.On("GetCurrentVersion", identifierBytes).Return(nil, fmt.Errorf("not found"))
	res, err = h.Get(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.Nil(t, res)
	assert.EqualError(t, err, "document not found: not found")
}

func TestGrpcHandler_Get(t *testing.T) {
	identifier := "0x01010101"
	identifierBytes, _ := hexutil.Decode(identifier)
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientinvoicepb.GetRequest{Identifier: identifier}
	response := &clientinvoicepb.InvoiceResponse{}
	srv.On("GetCurrentVersion", identifierBytes).Return(model, nil)
	srv.On("DeriveInvoiceResponse", model).Return(response, nil)
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
	payload := &clientinvoicepb.GetVersionRequest{Identifier: "0x0x", Version: "0x00"}
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
	payload := &clientinvoicepb.GetVersionRequest{Identifier: "0x01", Version: "0x00"}

	response := &clientinvoicepb.InvoiceResponse{}
	srv.On("GetVersion", []byte{0x01}, []byte{0x00}).Return(model, nil)
	srv.On("DeriveInvoiceResponse", model).Return(response, nil)
	res, err := h.GetVersion(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, response)
}

func TestGrpcHandler_Update_derive_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	payload := &clientinvoicepb.InvoiceUpdatePayload{Identifier: "0x010201"}
	srv.On("DeriveFromUpdatePayload", payload, mock.Anything).Return(nil, fmt.Errorf("derive error")).Once()
	res, err := h.Update(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive error")
	assert.Nil(t, res)
}

func TestGrpcHandler_Update_update_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := &mockModel{}
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)
	payload := &clientinvoicepb.InvoiceUpdatePayload{Identifier: "0x010201"}
	srv.On("DeriveFromUpdatePayload", payload, mock.Anything).Return(model, nil).Once()
	srv.On("Update", ctxh, model).Return(nil, fmt.Errorf("update error")).Once()
	res, err := h.Update(ctx, payload)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
	assert.Nil(t, res)
}

func TestGrpcHandler_Update_derive_response_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := &mockModel{}
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)
	payload := &clientinvoicepb.InvoiceUpdatePayload{Identifier: "0x010201"}
	srv.On("DeriveFromUpdatePayload", payload, mock.Anything).Return(model, nil).Once()
	srv.On("Update", ctxh, model).Return(model, nil).Once()
	srv.On("DeriveInvoiceResponse", model).Return(nil, fmt.Errorf("derive response error")).Once()
	res, err := h.Update(ctx, payload)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive response error")
	assert.Nil(t, res)
}

func TestGrpcHandler_Update(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := &mockModel{}
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)
	payload := &clientinvoicepb.InvoiceUpdatePayload{Identifier: "0x010201"}
	resp := &clientinvoicepb.InvoiceResponse{}
	srv.On("DeriveFromUpdatePayload", payload, mock.Anything).Return(model, nil).Once()
	srv.On("Update", ctxh, model).Return(model, nil).Once()
	srv.On("DeriveInvoiceResponse", model).Return(resp, nil).Once()
	res, err := h.Update(ctx, payload)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, resp, res)
}
