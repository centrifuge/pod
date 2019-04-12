// +build unit integration

package testingdocuments

import (
	"context"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/identity"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/mock"

)


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
func (m *MockEntityRelationService) DeriveFromCreatePayload(ctx context.Context, payload *entitypb2.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(documents.Model), args.Error(1)
}

// DeriveFromUpdatePayload derives a revoked entity relationship model from RelationshipPayload
func (m *MockEntityRelationService) DeriveFromUpdatePayload(ctx context.Context, payload *entitypb2.RelationshipPayload) (documents.Model, error){
	args := m.Called(ctx, payload)
	return args.Get(0).(documents.Model), args.Error(1)
}

// DeriveEntityRelationshipData returns the entity relationship data as client data
func (m *MockEntityRelationService) DeriveEntityRelationshipData(relationship documents.Model) (*entitypb2.RelationshipData, error){
	args := m.Called(relationship)
	return args.Get(0).(*entitypb2.RelationshipData), args.Error(1)
}

// DeriveEntityRelationshipResponse returns the entity relationship model in our standard client format
func (m *MockEntityRelationService) DeriveEntityRelationshipResponse(relationship documents.Model) (*entitypb2.RelationshipResponse, error){
	args := m.Called(relationship)
	return args.Get(0).(*entitypb2.RelationshipResponse), args.Error(1)
}

// GetEntityRelationships returns a list of the latest versions of the relevant entity relationship based on an entity id
func  (m *MockEntityRelationService) GetEntityRelationships(ctx context.Context, entityID []byte) ([]entityrelationship.EntityRelationship, error){
	args := m.Called(ctx, entityID)
	return args.Get(0).([]entityrelationship.EntityRelationship), args.Error(1)
}




func CreateRelationship() *entitypb.EntityRelationship {
	did, _ := identity.StringsToDIDs("0xed03Fa80291fF5DDC284DE6b51E716B130b05e20", "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	return &entitypb.EntityRelationship{
		OwnerIdentity:  did[0][:],
		TargetIdentity: did[1][:],
	}
}

func CreateRelationshipPayload() *entitypb2.RelationshipPayload {
	did2 := "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb"
	entityID := hexutil.Encode(utils.RandomSlice(32))

	return &entitypb2.RelationshipPayload{
		TargetIdentity: did2,
		Identifier:     entityID,
	}
}
