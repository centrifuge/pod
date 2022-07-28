//go:build integration || unit || testworld
// +build integration unit testworld

package entity

import (
	"context"
	"encoding/json"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/stretchr/testify/assert"
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

func (m *MockEntityRelationService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Document, error) {
	args := m.Called(documentID)
	return args.Get(0).(documents.Document), args.Error(1)
}

func (m *MockEntityRelationService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Document, error) {
	args := m.Called(documentID, version)
	return args.Get(0).(documents.Document), args.Error(1)
}

func (m *MockEntityRelationService) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(documentID, fields)
	return args.Get(0).(*documents.DocumentProof), args.Error(1)
}

func (m *MockEntityRelationService) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(documentID, version, fields)
	return args.Get(0).(*documents.DocumentProof), args.Error(1)
}

func (m *MockEntityRelationService) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Document, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Document), args.Error(1)
}

func (m *MockEntityRelationService) RequestDocumentSignature(ctx context.Context, model documents.Document, collaborator identity.DID) ([]*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).([]*coredocumentpb.Signature), args.Error(1)
}

func (m *MockEntityRelationService) ReceiveAnchoredDocument(ctx context.Context, model documents.Document, collaborator identity.DID) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEntityRelationService) Exists(ctx context.Context, documentID []byte) bool {
	args := m.Called()
	return args.Get(0).(bool)
}

// GetEntityRelationships returns a list of the latest versions of the relevant entity relationship based on an entity id
func (m *MockEntityRelationService) GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Document, error) {
	args := m.Called(ctx, entityID)
	rs, _ := args.Get(0).([]documents.Document)
	return rs, args.Error(1)
}

type MockService struct {
	Service
	mock.Mock
}

func (m *MockService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Document, error) {
	args := m.Called(ctx, documentID)
	data, _ := args.Get(0).(documents.Document)
	return data, args.Error(1)
}

func (m *MockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Document, error) {
	args := m.Called(ctx, documentID, version)
	data, _ := args.Get(0).(documents.Document)
	return data, args.Error(1)
}

func (m *MockService) GetEntityByRelationship(ctx context.Context, rID []byte) (documents.Document, error) {
	args := m.Called(ctx, rID)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func entityData(t *testing.T) []byte {
	did, err := identity.NewDIDFromString("0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.NoError(t, err)
	data := Data{
		Identity:  &did,
		LegalName: "Hello, world",
		Addresses: []Address{
			{
				Country: "Germany",
				IsMain:  true,
				Label:   "office",
			},
		},
		Contacts: []Contact{
			{
				Name:  "John Doe",
				Title: "Mr",
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	return d
}

func CreateEntityPayload(t *testing.T, collaborators []identity.DID) documents.CreatePayload {
	//if collaborators == nil {
	//	collaborators = []identity.DID{testingidentity.GenerateRandomDID()}
	//}
	return documents.CreatePayload{
		Scheme: Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: collaborators,
		},
		Data: entityData(t),
	}
}

func InitEntity(t *testing.T, did identity.DID, payload documents.CreatePayload) *Entity {
	entity := new(Entity)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	assert.NoError(t, entity.DeriveFromCreatePayload(context.Background(), payload))
	return entity
}

func CreateEntityWithEmbedCDWithPayload(t *testing.T, ctx context.Context, did identity.DID, payload documents.CreatePayload) (*Entity, coredocumentpb.CoreDocument) {
	entity := new(Entity)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	err := entity.DeriveFromCreatePayload(ctx, payload)
	assert.NoError(t, err)
	entity.GetTestCoreDocWithReset()
	sr, err := entity.CalculateSigningRoot()
	assert.NoError(t, err)
	// if acc errors out, just skip it
	if ctx == nil {
		ctx = context.Background()
	}
	acc, err := contextutil.Account(ctx)
	if err == nil {
		sig, err := acc.SignMsg(sr)
		assert.NoError(t, err)
		entity.AppendSignatures(sig)
	}
	_, err = entity.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)
	return entity, cd
}

func CreateEntityWithEmbedCD(t *testing.T, ctx context.Context, did identity.DID, collaborators []identity.DID) (*Entity, coredocumentpb.CoreDocument) {
	payload := CreateEntityPayload(t, collaborators)
	return CreateEntityWithEmbedCDWithPayload(t, ctx, did, payload)
}

// unpackFromUpdatePayload unpacks the update payload and prepares a new version.
func (e *Entity) unpackFromUpdatePayload(old *Entity, payload documents.UpdatePayload) error {
	var d Data
	if err := loadData(payload.Data, &d); err != nil {
		return errors.NewTypedError(ErrEntityInvalidData, err)
	}

	ncd, err := old.CoreDocument.PrepareNewVersion(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	e.Data = d
	e.CoreDocument = ncd
	return nil
}
