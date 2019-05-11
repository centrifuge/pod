// +build unit

package entity

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/anchors"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
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
func (m *MockEntityRelationService) DeriveFromUpdatePayload(ctx context.Context, payload *entitypb2.RelationshipPayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(documents.Model), args.Error(1)
}

// DeriveEntityRelationshipData returns the entity relationship data as client data
func (m *MockEntityRelationService) DeriveEntityRelationshipData(relationship documents.Model) (*entitypb2.RelationshipData, error) {
	args := m.Called(relationship)
	return args.Get(0).(*entitypb2.RelationshipData), args.Error(1)
}

// DeriveEntityRelationshipResponse returns the entity relationship model in our standard client format
func (m *MockEntityRelationService) DeriveEntityRelationshipResponse(relationship documents.Model) (*entitypb2.RelationshipResponse, error) {
	args := m.Called(relationship)
	return args.Get(0).(*entitypb2.RelationshipResponse), args.Error(1)
}

// GetEntityRelationships returns a list of the latest versions of the relevant entity relationship based on an entity id
func (m *MockEntityRelationService) GetEntityRelationships(ctx context.Context, entityID []byte) ([]documents.Model, error) {
	args := m.Called(ctx, entityID)
	return args.Get(0).([]documents.Model), args.Error(1)
}

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, *testingcommons.MockIdentityFactory, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(dIDBytes, nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	idFactory := new(testingcommons.MockIdentityFactory)
	repo := testRepo()
	mockAnchor := &mockAnchorRepo{}
	docSrv := documents.DefaultService(cfg, repo, mockAnchor, documents.NewServiceRegistry(), &idService)
	anchorRepo := &testinganchors.MockAnchorRepo{}
	anchorRepo.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	return idService, idFactory, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[jobs.BootstrappedService].(jobs.Manager), idFactory,
		nil, nil, anchorRepo, nil, nil)
}

func TestService_Update(t *testing.T) {
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// missing last version
	model, _ := createCDWithEmbeddedEntity(t)
	_, _, _, err := eSrv.Update(ctxh, model)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.NoError(t, testRepo().Create(accountID, model.CurrentVersion(), model))

	// calculate data root fails
	nm := new(mockModel)
	nm.On("ID").Return(model.ID(), nil).Once()
	_, _, _, err = eSrv.Update(ctxh, nm)
	nm.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	data, err := eSrv.DeriveEntityData(model)
	assert.NoError(t, err)
	data.LegalName = "test company"
	data.Contacts = []*entitypb.Contact{{Name: "Mr. Test"}}
	collab := testingidentity.GenerateRandomDID().String()
	newInv, err := eSrv.DeriveFromUpdatePayload(ctxh, &cliententitypb.EntityUpdatePayload{
		Identifier:  hexutil.Encode(model.ID()),
		WriteAccess: &documentpb.WriteAccess{Collaborators: []string{collab}},
		Data:        data,
	})
	assert.NoError(t, err)
	newData, err := eSrv.DeriveEntityData(newInv)
	assert.NoError(t, err)
	assert.Equal(t, data, newData)

	model, _, _, err = eSrv.Update(ctxh, newInv)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.True(t, testRepo().Exists(accountID, model.ID()))
	assert.True(t, testRepo().Exists(accountID, model.CurrentVersion()))
	assert.True(t, testRepo().Exists(accountID, model.PreviousVersion()))

	newData, err = eSrv.DeriveEntityData(model)
	assert.NoError(t, err)
	assert.Equal(t, data, newData)
	idFactory.AssertExpectations(t)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()
	// nil payload
	doc, err := eSrv.DeriveFromUpdatePayload(nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrPayloadNil, err))
	assert.Nil(t, doc)

	// nil payload data
	doc, err = eSrv.DeriveFromUpdatePayload(nil, &cliententitypb.EntityUpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrPayloadNil, err))
	assert.Nil(t, doc)

	// messed up identifier
	contextHeader := testingconfig.CreateAccountContext(t, cfg)
	payload := &cliententitypb.EntityUpdatePayload{Identifier: "some identifier", Data: &cliententitypb.EntityData{}}
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentIdentifier, err))
	assert.Contains(t, err.Error(), "failed to decode identifier")
	assert.Nil(t, doc)

	// missing last version
	id := utils.RandomSlice(32)
	payload.Identifier = hexutil.Encode(id)
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.Nil(t, doc)

	// Entity data does not contain an identity
	old, _ := createCDWithEmbeddedEntity(t)
	err = testRepo().Create(accountID, old.CurrentVersion(), old)
	assert.NoError(t, err)
	payload.Data = &cliententitypb.EntityData{
		LegalName: "test company",
		Contacts:  []*entitypb.Contact{{Name: "Mr. Test"}},
	}

	payload.Identifier = hexutil.Encode(old.ID())
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)

	assert.Error(t, err, "should fail because Identity is missing")
	assert.Nil(t, doc)

	// invalid collaborator identity
	payload.Data.LegalName = "new company name"
	payload.WriteAccess = &documentpb.WriteAccess{Collaborators: []string{"some wrong ID"}}
	payload.Data.Identity = testingidentity.GenerateRandomDID().String()
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.Nil(t, doc)

	// success
	wantCollab := testingidentity.GenerateRandomDID()
	payload.WriteAccess = &documentpb.WriteAccess{Collaborators: []string{wantCollab.String()}}
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
	cs, err := doc.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs.ReadWriteCollaborators, 3)
	assert.Contains(t, cs.ReadWriteCollaborators, wantCollab)
	assert.Equal(t, old.ID(), doc.ID())
	assert.Equal(t, payload.Identifier, hexutil.Encode(doc.ID()))
	assert.Equal(t, old.CurrentVersion(), doc.PreviousVersion())
	assert.Equal(t, old.NextVersion(), doc.CurrentVersion())
	assert.NotNil(t, doc.NextVersion())
	data, err := doc.(*Entity).getClientData()
	assert.NoError(t, err)
	assert.Equal(t, payload.Data, data)
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	eSrv := service{}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	m, err := eSrv.DeriveFromCreatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrPayloadNil, err))

	// nil data payload
	m, err = eSrv.DeriveFromCreatePayload(ctxh, &cliententitypb.EntityCreatePayload{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrPayloadNil, err))

	// Init fails
	payload := &cliententitypb.EntityCreatePayload{
		Data: &cliententitypb.EntityData{
			Identity:  "i am not a did",
			LegalName: "company test",
		},
	}

	m, err = eSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.Data.Identity = testingidentity.GenerateRandomDID().String()
	m, err = eSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	entity := m.(*Entity)
	assert.Equal(t, entity.LegalName, payload.Data.LegalName)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	eSrv := service{repo: testRepo()}
	_, cd := createCDWithEmbeddedEntity(t)
	m, err := eSrv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	entity, ok := m.(*Entity)
	assert.True(t, ok, "must be true")
	assert.Equal(t, entity.LegalName, "Company Test")
	assert.Equal(t, entity.Contacts[0].Name, "Satoshi Nakamoto")
}

func TestService_Create(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// calculate data root fails
	m, _, _, err := eSrv.Create(ctxh, &mockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	entity, err := eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.NoError(t, err)
	m, _, _, err = eSrv.Create(ctxh, entity)
	assert.NoError(t, err)
	assert.True(t, testRepo().Exists(accountID, m.ID()))
	assert.True(t, testRepo().Exists(accountID, m.CurrentVersion()))
	idFactory.AssertExpectations(t)
}

func TestService_DeriveEntityData(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()

	// some random model
	_, err := eSrv.DeriveEntityData(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := testingdocuments.CreateEntityPayload()
	entity, err := eSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.NoError(t, err, "must be non nil")
	data, err := eSrv.DeriveEntityData(entity)
	assert.NoError(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
}

func TestService_DeriveEntityResponse(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// prepare a service with mocked layers
	ctxh, entity, er, idFactory, idService, repo := setupRelationshipTesting(t)
	eID := entity.ID()
	erID := er.ID()

	// testcase: request from peer
	mockAnchor := &mockAnchorRepo{}
	docSrv := testingdocuments.MockService{}
	mockedERSrv := &MockEntityRelationService{}
	mockProcessor := &testingcommons.MockRequestProcessor{}

	docSrv.On("GetCurrentVersion", eID).Return(entity, nil)
	docSrv.On("Exists").Return(true).Once()
	mockedERSrv.On("GetCurrentVersion", er.ID()).Return(er, nil)

	fakeRoot, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)
	nextId, err := anchors.ToAnchorID(entity.NextVersion())
	assert.NoError(t, err)
	mockAnchor.On("GetAnchorData", nextId).Return(fakeRoot, time.Now(), nil).Once()

	token, err := er.GetAccessTokens()
	assert.NoError(t, err)

	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)

	mockProcessor.On("RequestDocumentWithAccessToken", did, token[0].Identifier, eID, erID).Return(&p2ppb.GetDocumentResponse{Document: &cd}, nil)
	docSrv.On("DeriveFromCoreDocument", mock.Anything).Return(entity, nil)
	docSrv.On("Exists").Return(false).Once()
	mockedERSrv.On("GetEntityRelationships", mock.Anything, entity.ID()).Return([]documents.Model{er}, nil)
	//initialize service
	entitySrv := DefaultService(
		&docSrv,
		repo,
		nil,
		nil, idFactory,
		mockedERSrv, idService, mockAnchor, mockProcessor, func() documents.ValidatorGroup {
			return documents.ValidatorGroup{}
		})

	// derive data failed
	m := new(mockModel)
	r, err := entitySrv.DeriveEntityResponse(ctxh, m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	r, err = entitySrv.DeriveEntityResponse(ctxh, entity)
	assert.NoError(t, err)
	payload := testingdocuments.CreateEntityPayload()
	assert.Equal(t, payload.Data.Contacts[0].Name, r.Data.Entity.Contacts[0].Name)
	assert.Equal(t, payload.Data.LegalName, r.Data.Entity.LegalName)
	assert.Contains(t, r.Header.WriteAccess.Collaborators, did.String())

	// entity is not collaborator on document
	e := new(Entity)
	err = e.InitEntityInput(testingdocuments.CreateEntityPayload(), testingidentity.GenerateRandomDID())
	assert.NoError(t, err)
	_, err = e.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err = e.PackCoreDocument()
	assert.NoError(t, err)
	empty, err := entitySrv.DeriveEntityResponse(ctxh, e)
	assert.NoError(t, err)
	assert.Equal(t, empty.Data.Relationships, []*entitypb2.Relationship(nil))
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()
	doc, _ := createCDWithEmbeddedEntity(t)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.NoError(t, err)

	data, err := doc.(*Entity).getClientData()
	assert.NoError(t, err)
	data.LegalName = "test company"
	doc2 := new(Entity)
	assert.NoError(t, doc2.PrepareNewVersion(doc, data, documents.CollaboratorsAccess{}))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := eSrv.GetCurrentVersion(ctxh, doc.ID())

	doc3Entity := doc3.(*Entity)

	assert.NoError(t, err)
	assert.Equal(t, doc2.LegalName, doc3Entity.LegalName)
}

func TestService_GetVersion(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()
	entity, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := eSrv.GetVersion(ctxh, entity.ID(), entity.CurrentVersion())
	assert.NoError(t, err)

	mod, err = eSrv.GetVersion(ctxh, mod.ID(), []byte{})
	assert.Error(t, err)
}

func TestService_Get_Collaborators(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()
	entity, _ := createCDWithEmbeddedEntity(t)

	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)

	_, err = eSrv.GetVersion(ctxh, entity.ID(), entity.CurrentVersion())
	assert.NoError(t, err)

	// set other DID for selfDID
	oldDIDBytes, err := cfg.GetIdentityID()
	assert.NoError(t, err)
	oldDID, err := identity.NewDIDFromBytes(oldDIDBytes)
	assert.NoError(t, err)

	cfg.Set("identityId", testingidentity.GenerateRandomDID().ToAddress().String())
	ctxh = testingconfig.CreateAccountContext(t, cfg)

	_, err = eSrv.GetVersion(ctxh, entity.ID(), entity.CurrentVersion())

	//todo should currently fail because not implemented
	assert.Error(t, err)

	// reset to old DID for other test cases
	cfg.Set("identityId", oldDID.ToAddress().String())
}

func TestService_Exists(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()
	entity, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := eSrv.Exists(ctxh, entity.CurrentVersion())
	assert.True(t, exists, "entity should exist")

	exists = eSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, " entity should not exist")
}

func TestService_calculateDataRoot(t *testing.T) {
	eSrv := service{repo: testRepo()}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// type mismatch
	entity, err := eSrv.validateAndPersist(ctxh, nil, &mockModel{}, nil)
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	entity, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.NoError(t, err)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, v)
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	entity, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.NoError(t, err)
	err = eSrv.repo.Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)
	idFactory := new(testingcommons.MockIdentityFactory)
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil).Once()
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, CreateValidator(idFactory))
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)
	idFactory.AssertExpectations(t)

	// success
	idFactory = new(testingcommons.MockIdentityFactory)
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil).Once()
	entity, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.NoError(t, err)
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, CreateValidator(idFactory))
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	idFactory.AssertExpectations(t)
}

func setupRelationshipTesting(t *testing.T) (context.Context, documents.Model, *entityrelationship.EntityRelationship, identity.Factory, identity.Service, documents.Repository) {
	idService := &testingcommons.MockIdentityService{}
	idFactory := new(testingcommons.MockIdentityFactory)
	repo := testRepo()

	// successful request
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// create entity
	entity, _ := createCDWithEmbeddedEntity(t)

	// create relationship
	erData := &entitypb2.RelationshipData{
		EntityIdentifier: hexutil.Encode(entity.ID()),
		OwnerIdentity:    hexutil.Encode(dIDBytes),
		TargetIdentity:   hexutil.Encode(dIDBytes),
	}
	er := &entityrelationship.EntityRelationship{}
	err := er.InitEntityRelationshipInput(ctxh, hexutil.Encode(entity.ID()), erData)
	assert.NoError(t, err)

	return ctxh, entity, er, idFactory, idService, repo

}

// todo entity currently not stored in db
//func TestService_GetEntityByRelationship_latestInDB(t *testing.T) {
//	// prepare a service with mocked layers
//	ctxh, entity, er, idFactory, idService, repo := setupRelationshipTesting(t)
//
//	eID := entity.ID()
//	erID := er.ID()
//
//	// testcase: latest version in db
//	mockAnchor := &mockAnchorRepo{}
//	docSrv := testingdocuments.MockService{}
//	mockedERSrv := &MockEntityRelationService{}
//	mockProcessor := &testingcommons.MockRequestProcessor{}
//
//	docSrv.On("GetCurrentVersion", eID).Return(entity, nil)
//	docSrv.On("Exists").Return(true)
//	mockedERSrv.On("GetCurrentVersion", er.ID()).Return(er, nil)
//
//	zeros := [32]byte{}
//	zeroRoot, err := anchors.ToDocumentRoot(zeros[:])
//	nextId, err := anchors.ToAnchorID(entity.NextVersion())
//	mockAnchor.On("GetAnchorData", nextId).Return(zeroRoot, time.Now(), nil).Once()
//
//	//initialize service
//	entitySrv := DefaultService(
//		&docSrv,
//		repo,
//		nil,
//		nil, idFactory,
//		mockedERSrv, idService, mockAnchor, mockProcessor, nil)
//
//	//successful latest version in db
//	model, err := entitySrv.GetEntityByRelationship(ctxh, erID)
//	assert.NoError(t, err)
//	assert.Equal(t, model.CurrentVersion(), entity.CurrentVersion())
//
//}

func TestService_GetEntityByRelationship_fail(t *testing.T) {
	// prepare a service with mocked layers
	ctxh, entity, er, idFactory, idService, repo := setupRelationshipTesting(t)

	mockAnchor := &mockAnchorRepo{}
	docSrv := testingdocuments.MockService{}
	mockedERSrv := &MockEntityRelationService{}
	mockProcessor := &testingcommons.MockRequestProcessor{}

	mockedERSrv.On("GetCurrentVersion", er.ID()).Return(er, entityrelationship.ErrERNotFound)

	//initialize service
	entitySrv := DefaultService(
		&docSrv,
		repo,
		nil,
		nil, idFactory,
		mockedERSrv, idService, mockAnchor, mockProcessor, nil)

	//entity relationship identifier not exists in db
	model, err := entitySrv.GetEntityByRelationship(ctxh, er.ID())
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, err, entityrelationship.ErrERNotFound)

	//pass entity id instead of er identifier
	mockedERSrv.On("GetCurrentVersion", entity.ID()).Return(entity, nil)

	//initialize service
	entitySrv = DefaultService(
		&docSrv,
		repo,
		nil,
		nil, idFactory,
		mockedERSrv, idService, mockAnchor, mockProcessor, nil)

	// pass entity id instead of er identifier
	model, err = entitySrv.GetEntityByRelationship(ctxh, entity.ID())
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, err, entityrelationship.ErrNotEntityRelationship)

}

func TestService_GetEntityByRelationship_requestP2P(t *testing.T) {
	// prepare a service with mocked layers
	ctxh, entity, er, idFactory, idService, repo := setupRelationshipTesting(t)

	eID := entity.ID()
	erID := er.ID()

	// testcase: request from peer
	mockAnchor := &mockAnchorRepo{}
	docSrv := testingdocuments.MockService{}
	mockedERSrv := &MockEntityRelationService{}
	mockProcessor := &testingcommons.MockRequestProcessor{}

	docSrv.On("GetCurrentVersion", eID).Return(entity, nil)
	docSrv.On("Exists").Return(true).Once()
	mockedERSrv.On("GetCurrentVersion", er.ID()).Return(er, nil)

	fakeRoot, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)
	nextId, err := anchors.ToAnchorID(entity.NextVersion())
	assert.NoError(t, err)
	mockAnchor.On("GetAnchorData", nextId).Return(fakeRoot, time.Now(), nil).Once()

	token, err := er.GetAccessTokens()
	assert.NoError(t, err)

	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)

	mockProcessor.On("RequestDocumentWithAccessToken", did, token[0].Identifier, eID, erID).Return(&p2ppb.GetDocumentResponse{Document: &cd}, nil)
	docSrv.On("DeriveFromCoreDocument", mock.Anything).Return(entity, nil)
	docSrv.On("Exists").Return(false).Once()

	//initialize service
	entitySrv := DefaultService(
		&docSrv,
		repo,
		nil,
		nil, idFactory,
		mockedERSrv, idService, mockAnchor, mockProcessor, func() documents.ValidatorGroup {
			return documents.ValidatorGroup{}
		})

	//entity relationship is not the latest request therefore request from peer
	model, err := entitySrv.GetEntityByRelationship(ctxh, erID)
	assert.NoError(t, err)
	assert.Equal(t, model.CurrentVersion(), entity.CurrentVersion())

}
