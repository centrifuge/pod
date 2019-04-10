// +build unit

package entity

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	Service
	mock.Mock
}

func (m *mockService) DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityCreatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *mockService) DeriveFromSharePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *mockService) Create(ctx context.Context, model documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.TX(ctx), nil, args.Error(2)
}

func (m *mockService) Share(ctx context.Context, model documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.TX(ctx), nil, args.Error(2)
}

func (m *mockService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	args := m.Called(ctx, documentID)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *mockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	args := m.Called(ctx, documentID, version)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *mockService) DeriveEntityData(doc documents.Model) (*cliententitypb.EntityData, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*cliententitypb.EntityData)
	return data, args.Error(1)
}

func (m *mockService) DeriveEntityResponse(doc documents.Model) (*cliententitypb.EntityResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*cliententitypb.EntityResponse)
	return data, args.Error(1)
}

func (m *mockService) DeriveEntityRelationshipResponse(doc documents.Model) (*cliententitypb.RelationshipResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*cliententitypb.RelationshipResponse)
	return data, args.Error(1)
}

func (m *mockService) Update(ctx context.Context, model documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.TX(ctx), nil, args.Error(2)
}

func (m *mockService) Revoke(ctx context.Context, model documents.Model) (documents.Model, transactions.TxID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.TX(ctx), nil, args.Error(2)
}

func (m *mockService) DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityUpdatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func (m *mockService) DeriveFromRevokePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func getHandler() *grpcHandler {
	return &grpcHandler{service: &mockService{}, config: configService}
}

func TestGRPCHandler_Create_derive_fail(t *testing.T) {
	// DeriveFrom payload fails
	h := getHandler()
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(nil, errors.New("derive failed")).Once()
	_, err := h.Create(testingconfig.HandlerContext(configService), nil)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "derive failed")
}

func TestGRPCHandler_Create_create_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(new(Entity), nil).Once()
	srv.On("Create", mock.Anything, mock.Anything).Return(nil, transactions.NilTxID().String(), errors.New("create failed")).Once()
	payload := &cliententitypb.EntityCreatePayload{Data: &cliententitypb.EntityData{LegalName: "test company"}}
	_, err := h.Create(testingconfig.HandlerContext(configService), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "create failed")
}

func TestGRPCHandler_Create_DeriveEntityResponse_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(Entity)
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("Create", mock.Anything, mock.Anything).Return(model, transactions.NilTxID().String(), nil).Once()
	srv.On("DeriveEntityResponse", mock.Anything).Return(nil, errors.New("derive response failed"))
	payload := &cliententitypb.EntityCreatePayload{Data: &cliententitypb.EntityData{LegalName: "test company"}}
	_, err := h.Create(testingconfig.HandlerContext(configService), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "derive response failed")
}

func TestGrpcHandler_Create(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(Entity)
	txID := transactions.NewTxID()
	payload := &cliententitypb.EntityCreatePayload{
		Data:        &cliententitypb.EntityData{LegalName: "test company"},
		WriteAccess: &documentpb.WriteAccess{Collaborators: []string{"0x010203040506"}},
	}
	response := &cliententitypb.EntityResponse{Header: &documentpb.ResponseHeader{}}
	srv.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("Create", mock.Anything, mock.Anything).Return(model, txID.String(), nil).Once()
	srv.On("DeriveEntityResponse", model).Return(response, nil)
	res, err := h.Create(testingconfig.HandlerContext(configService), payload)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}

func TestGrpcHandler_Share(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := &mockModel{}
	txID := transactions.NewTxID()
	payload := &cliententitypb.RelationshipPayload{
		Identifier:     "0x010203",
		TargetIdentity: "some DID",
	}
	response := &cliententitypb.RelationshipResponse{Header: &documentpb.ResponseHeader{}}
	srv.On("DeriveFromSharePayload", mock.Anything, mock.Anything).Return(model, nil).Once()
	srv.On("Share", mock.Anything, mock.Anything).Return(model, txID.String(), nil).Once()
	srv.On("DeriveEntityRelationshipResponse", model).Return(response, nil)
	res, err := h.Share(testingconfig.HandlerContext(configService), payload)
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
	payload := &cliententitypb.GetRequest{Identifier: "invalid"}

	res, err := h.Get(testingconfig.HandlerContext(configService), payload)
	assert.Nil(t, res)
	assert.EqualError(t, err, "identifier is an invalid hex string: hex string without 0x prefix")

	payload.Identifier = identifier
	srv.On("GetCurrentVersion", mock.Anything, identifierBytes).Return(nil, errors.New("not found"))
	res, err = h.Get(testingconfig.HandlerContext(configService), payload)
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
	payload := &cliententitypb.GetRequest{Identifier: identifier}
	response := &cliententitypb.EntityResponse{}
	srv.On("GetCurrentVersion", mock.Anything, identifierBytes).Return(model, nil)
	srv.On("DeriveEntityResponse", model).Return(response, nil)
	res, err := h.Get(testingconfig.HandlerContext(configService), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}

func TestGrpcHandler_GetVersion_invalid_input(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	payload := &cliententitypb.GetVersionRequest{Identifier: "0x0x", Version: "0x00"}
	res, err := h.GetVersion(testingconfig.HandlerContext(configService), payload)
	assert.EqualError(t, err, "identifier is invalid: invalid hex string")
	payload.Version = "0x0x"
	payload.Identifier = "0x01"

	res, err = h.GetVersion(testingconfig.HandlerContext(configService), payload)
	assert.EqualError(t, err, "version is invalid: invalid hex string")
	payload.Version = "0x00"
	payload.Identifier = "0x01"

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
	payload := &cliententitypb.GetVersionRequest{Identifier: "0x01", Version: "0x00"}

	response := &cliententitypb.EntityResponse{}
	srv.On("GetVersion", mock.Anything, []byte{0x01}, []byte{0x00}).Return(model, nil)
	srv.On("DeriveEntityResponse", model).Return(response, nil)
	res, err := h.GetVersion(testingconfig.HandlerContext(configService), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, response)
}

func TestGrpcHandler_Update_derive_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	payload := &cliententitypb.EntityUpdatePayload{Identifier: "0x010201"}
	srv.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(nil, errors.New("derive error")).Once()
	res, err := h.Update(testingconfig.HandlerContext(configService), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive error")
	assert.Nil(t, res)
}

func TestGrpcHandler_Update_update_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := &mockModel{}
	ctx := testingconfig.HandlerContext(configService)
	payload := &cliententitypb.EntityUpdatePayload{Identifier: "0x010201"}
	srv.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(model, nil).Once()
	srv.On("Update", mock.Anything, model).Return(nil, transactions.NilTxID().String(), errors.New("update error")).Once()
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
	ctx := testingconfig.HandlerContext(configService)
	payload := &cliententitypb.EntityUpdatePayload{Identifier: "0x010201"}
	srv.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(model, nil).Once()
	srv.On("Update", mock.Anything, model).Return(model, transactions.NilTxID().String(), nil).Once()
	srv.On("DeriveEntityResponse", model).Return(nil, errors.New("derive response error")).Once()
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
	ctx := testingconfig.HandlerContext(configService)
	txID := transactions.NewTxID()
	payload := &cliententitypb.EntityUpdatePayload{Identifier: "0x010201"}
	resp := &cliententitypb.EntityResponse{Header: new(documentpb.ResponseHeader)}
	srv.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(model, nil).Once()
	srv.On("Update", mock.Anything, model).Return(model, txID.String(), nil).Once()
	srv.On("DeriveEntityResponse", model).Return(resp, nil).Once()
	res, err := h.Update(ctx, payload)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, resp, res)
}

func TestGrpcHandler_Revoke(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := &mockModel{}
	ctx := testingconfig.HandlerContext(configService)
	txID := transactions.NewTxID()
	payload := &cliententitypb.RelationshipPayload{Identifier: "0x010201", TargetIdentity: "some DID"}
	resp := &cliententitypb.RelationshipResponse{Header: new(documentpb.ResponseHeader)}
	srv.On("DeriveFromRevokePayload", mock.Anything, payload).Return(model, nil).Once()
	srv.On("Revoke", mock.Anything, model).Return(model, txID.String(), nil).Once()
	srv.On("DeriveEntityRelationshipResponse", model).Return(resp, nil).Once()
	res, err := h.Revoke(ctx, payload)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, resp, res)
}
