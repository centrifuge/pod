// +build unit

package entity

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
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

func (m *mockService) Create(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *mockService) Share(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.Job(ctx), nil, args.Error(2)
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

func (m *mockService) DeriveEntityResponse(ctx context.Context, doc documents.Model) (*cliententitypb.EntityResponse, error) {
	args := m.Called(ctx, doc)
	data, _ := args.Get(0).(*cliententitypb.EntityResponse)
	return data, args.Error(1)
}

func (m *mockService) DeriveEntityRelationshipResponse(doc documents.Model) (*cliententitypb.RelationshipResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*cliententitypb.RelationshipResponse)
	return data, args.Error(1)
}

func (m *mockService) Update(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *mockService) Revoke(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.Job(ctx), nil, args.Error(2)
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
	srv.On("DeriveEntityResponse", mock.Anything, model).Return(response, nil)
	res, err := h.GetVersion(testingconfig.HandlerContext(configService), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res, response)
}

func TestGrpcHandler_Revoke(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	ctx := testingconfig.HandlerContext(configService)
	jobID := jobs.NewJobID()
	payload := &cliententitypb.RelationshipPayload{DocumentId: "0x010201", TargetIdentity: "some DID"}
	resp := &cliententitypb.RelationshipResponse{Header: new(documentpb.ResponseHeader)}
	srv.On("DeriveFromRevokePayload", mock.Anything, payload).Return(model, nil).Once()
	srv.On("Revoke", mock.Anything, model).Return(model, jobID.String(), nil).Once()
	srv.On("DeriveEntityRelationshipResponse", model).Return(resp, nil).Once()
	res, err := h.Revoke(ctx, payload)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, resp, res)
}
