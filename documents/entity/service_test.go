//go:build unit
// +build unit

package entity

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, *identity.MockFactory, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(dIDBytes, nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	idFactory := new(identity.MockFactory)
	repo := testRepo()
	anchorSrv := &anchors.MockAnchorService{}
	anchorSrv.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(cfg, repo, anchorSrv, documents.NewServiceRegistry(), &idService, nil)
	return idService, idFactory, DefaultService(
		docSrv,
		repo,
		nil, nil, anchorSrv, nil, nil)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	eSrv := service{repo: testRepo()}
	_, cd := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	m, err := eSrv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	entity, ok := m.(*Entity)
	assert.True(t, ok, "must be true")
	assert.Equal(t, entity.Data.LegalName, "Hello, world")
	assert.Equal(t, entity.Data.Contacts[0].Name, "John Doe")
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()
	doc, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	assert.NoError(t, doc.SetStatus(documents.Committed))
	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.NoError(t, err)

	data := doc.Data
	data.LegalName = "test company"
	doc2 := new(Entity)
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	err = doc2.unpackFromUpdatePayload(doc, documents.UpdatePayload{
		DocumentID: doc.ID(),
		CreatePayload: documents.CreatePayload{
			Data: d,
		},
	})
	assert.NoError(t, err)
	assert.NoError(t, doc2.SetStatus(documents.Committed))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := eSrv.GetCurrentVersion(ctxh, doc.ID())

	doc3Entity := doc3.(*Entity)

	assert.NoError(t, err)
	assert.Equal(t, doc2.Data.LegalName, doc3Entity.Data.LegalName)
}

func TestService_GetVersion(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, _, eSrv := getServiceWithMockedLayers()
	entity, _ := CreateEntityWithEmbedCD(t, ctxh, did, nil)
	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

	mod, err := eSrv.GetVersion(ctxh, entity.ID(), entity.CurrentVersion())
	assert.NoError(t, err)

	mod, err = eSrv.GetVersion(ctxh, mod.ID(), []byte{})
	assert.Error(t, err)
}

func TestService_Get_Collaborators(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, _, eSrv := getServiceWithMockedLayers()
	entity, _ := CreateEntityWithEmbedCD(t, ctxh, did, nil)

	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

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

	// todo should currently fail because not implemented
	assert.Error(t, err)

	// reset to old DID for other test cases
	cfg.Set("identityId", oldDID.ToAddress().String())
}

func setupRelationshipTesting(t *testing.T) (context.Context, documents.Document, *entityrelationship.EntityRelationship, identity.Factory, identity.Service, documents.Repository) {
	idService := &testingcommons.MockIdentityService{}
	idFactory := new(identity.MockFactory)
	repo := testRepo()

	// successful request
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// create entity
	entity, _ := CreateEntityWithEmbedCD(t, ctxh, did, nil)

	// create relationship
	erData := entityrelationship.Data{
		EntityIdentifier: entity.ID(),
		OwnerIdentity:    &did,
		TargetIdentity:   &did,
	}
	er := entityrelationship.InitEntityRelationship(t, ctxh, erData)
	return ctxh, entity, er, idFactory, idService, repo
}

func TestService_GetEntityByRelationship_fail(t *testing.T) {
	// prepare a service with mocked layers
	ctxh, entity, er, idFactory, _, repo := setupRelationshipTesting(t)

	mockAnchor := new(anchors.MockAnchorService)
	docSrv := documents.NewServiceMock(t)
	mockedERSrv := &MockEntityRelationService{}
	mockProcessor := &documents.MockRequestProcessor{}

	mockedERSrv.On("GetCurrentVersion", er.ID()).Return(er, entityrelationship.ErrERNotFound)

	// initialize service
	entitySrv := DefaultService(
		docSrv,
		repo,
		idFactory,
		mockedERSrv, mockAnchor, mockProcessor, nil)

	// entity relationship identifier not exists in db
	model, err := entitySrv.GetEntityByRelationship(ctxh, er.ID())
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, err, entityrelationship.ErrERNotFound)

	// pass entity id instead of er identifier
	mockedERSrv.On("GetCurrentVersion", entity.ID()).Return(entity, nil)

	// initialize service
	entitySrv = DefaultService(
		docSrv,
		repo,
		idFactory,
		mockedERSrv, mockAnchor, mockProcessor, nil)

	// pass entity id instead of er identifier
	model, err = entitySrv.GetEntityByRelationship(ctxh, entity.ID())
	assert.Error(t, err)
	assert.Nil(t, model)
	assert.Contains(t, err, entityrelationship.ErrNotEntityRelationship)

}

func TestService_GetEntityByRelationship_requestP2P(t *testing.T) {
	// prepare a service with mocked layers
	ctxh, entity, er, idFactory, _, repo := setupRelationshipTesting(t)

	eID := entity.ID()
	erID := er.ID()

	// testcase: request from peer
	mockAnchor := new(anchors.MockAnchorService)
	docSrv := documents.NewServiceMock(t)
	mockedERSrv := &MockEntityRelationService{}
	mockProcessor := &documents.MockRequestProcessor{}

	mockedERSrv.On("GetCurrentVersion", er.ID()).Return(er, nil)

	fakeRoot, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)
	nextID, err := anchors.ToAnchorID(entity.NextVersion())
	assert.NoError(t, err)
	mockAnchor.On("GetAnchorData", nextID).Return(fakeRoot, time.Now(), nil).Once()

	token, err := er.GetAccessTokens()
	assert.NoError(t, err)

	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)

	mockProcessor.On("RequestDocumentWithAccessToken", did, token[0].Identifier, eID, erID).Return(&p2ppb.GetDocumentResponse{Document: &cd}, nil)
	docSrv.On("DeriveFromCoreDocument", mock.Anything).Return(entity, nil)

	// initialize service
	entitySrv := DefaultService(
		docSrv,
		repo,
		idFactory,
		mockedERSrv,
		mockAnchor,
		mockProcessor,
		func() documents.ValidatorGroup {
			return documents.ValidatorGroup{}
		},
	)

	// entity relationship is not the latest request therefore request from peer
	model, err := entitySrv.GetEntityByRelationship(ctxh, erID)
	assert.NoError(t, err)
	assert.Equal(t, model.CurrentVersion(), entity.CurrentVersion())
}

func TestService_ValidateError(t *testing.T) {
	srv := service{}
	err := srv.Validate(context.Background(), nil, nil)
	assert.Error(t, err)
}
