// +build integration unit

package entity

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

type MockEntityRelationService struct {
	documents.Service
	mock.Mock
}

func (m *MockEntityRelationService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	args := m.Called(documentID)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockEntityRelationService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	args := m.Called(documentID, version)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockEntityRelationService) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(documentID, fields)
	return args.Get(0).(*documents.DocumentProof), args.Error(1)
}

func (m *MockEntityRelationService) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(documentID, version, fields)
	return args.Get(0).(*documents.DocumentProof), args.Error(1)
}

func (m *MockEntityRelationService) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Model, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockEntityRelationService) RequestDocumentSignature(ctx context.Context, model documents.Model, collaborator identity.DID) (*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).(*coredocumentpb.Signature), args.Error(1)
}

func (m *MockEntityRelationService) ReceiveAnchoredDocument(ctx context.Context, model documents.Model, collaborator identity.DID) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEntityRelationService) Exists(ctx context.Context, documentID []byte) bool {
	args := m.Called()
	return args.Get(0).(bool)
}

// DeriveFromCreatePayload derives Entity Relationship from RelationshipPayload
func (m *MockEntityRelationService) DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(documents.Model), args.Error(1)
}

// DeriveFromUpdatePayload derives a revoked entity relationship model from RelationshipPayload
func (m *MockEntityRelationService) DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(documents.Model), args.Error(1)
}

// DeriveEntityRelationshipData returns the entity relationship data as client data
func (m *MockEntityRelationService) DeriveEntityRelationshipData(relationship documents.Model) (*cliententitypb.RelationshipData, error) {
	args := m.Called(relationship)
	return args.Get(0).(*cliententitypb.RelationshipData), args.Error(1)
}

// DeriveEntityRelationshipResponse returns the entity relationship model in our standard client format
func (m *MockEntityRelationService) DeriveEntityRelationshipResponse(relationship documents.Model) (*cliententitypb.RelationshipResponse, error) {
	args := m.Called(relationship)
	return args.Get(0).(*cliententitypb.RelationshipResponse), args.Error(1)
}

// GetEntityRelationships returns a list of the latest versions of the relevant entity relationship based on an entity id
func (m *MockEntityRelationService) GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Model, error) {
	args := m.Called(ctx, entityID)
	rs, _ := args.Get(0).([]documents.Model)
	return rs, args.Error(1)
}

type MockService struct {
	Service
	mock.Mock
}

func (m *MockService) DeriveFromCreatePayload(ctx context.Context, payload *cliententitypb.EntityCreatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) DeriveFromSharePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) Create(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *MockService) Share(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	model, _ = args.Get(0).(documents.Model)
	return model, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *MockService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	args := m.Called(ctx, documentID)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *MockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	args := m.Called(ctx, documentID, version)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *MockService) DeriveEntityData(doc documents.Model) (*cliententitypb.EntityData, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*cliententitypb.EntityData)
	return data, args.Error(1)
}

func (m *MockService) DeriveEntityResponse(ctx context.Context, doc documents.Model) (*cliententitypb.EntityResponse, error) {
	args := m.Called(ctx, doc)
	data, _ := args.Get(0).(*cliententitypb.EntityResponse)
	return data, args.Error(1)
}

func (m *MockService) DeriveEntityRelationshipResponse(doc documents.Model) (*cliententitypb.RelationshipResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*cliententitypb.RelationshipResponse)
	return data, args.Error(1)
}

func (m *MockService) Update(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *MockService) Revoke(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, chan bool, error) {
	args := m.Called(ctx, model)
	doc1, _ := args.Get(0).(documents.Model)
	return doc1, contextutil.Job(ctx), nil, args.Error(2)
}

func (m *MockService) DeriveFromUpdatePayload(ctx context.Context, payload *cliententitypb.EntityUpdatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func (m *MockService) DeriveFromRevokePayload(ctx context.Context, payload *cliententitypb.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func (m *MockService) GetEntityByRelationship(ctx context.Context, rID []byte) (documents.Model, error) {
	args := m.Called(ctx, rID)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}
