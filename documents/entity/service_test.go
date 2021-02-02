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
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/anchors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, *identity.MockFactory, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(dIDBytes, nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	idFactory := new(identity.MockFactory)
	repo := testRepo()
	anchorSrv := &testinganchors.MockAnchorService{}
	anchorSrv.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(cfg, repo, anchorSrv, documents.NewServiceRegistry(), &idService, nil, nil, nil)
	return idService, idFactory, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[jobs.BootstrappedService].(jobs.Manager),
		nil, nil, anchorSrv, nil, nil)
}

func TestService_Update(t *testing.T) {
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// missing last version
	model, _ := CreateEntityWithEmbedCD(t, ctxh, did, nil)
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
	data := model.Data
	data.LegalName = "test company"
	data.Contacts = []Contact{{Name: "Mr. Test"}}
	collab := testingidentity.GenerateRandomDID()
	newEntity := new(Entity)
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	err = newEntity.unpackFromUpdatePayload(model, documents.UpdatePayload{
		DocumentID: model.ID(),
		CreatePayload: documents.CreatePayload{
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []identity.DID{collab},
			},
			Data: d,
		},
	})
	assert.NoError(t, err)
	newData := newEntity.Data
	assert.Equal(t, data, newData)

	eSrv.factory = idFactory
	doc, _, _, err := eSrv.Update(ctxh, newEntity)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.True(t, testRepo().Exists(accountID, doc.ID()))
	assert.True(t, testRepo().Exists(accountID, doc.CurrentVersion()))
	assert.True(t, testRepo().Exists(accountID, doc.PreviousVersion()))

	newData = doc.GetData().(Data)
	assert.NoError(t, err)
	assert.Equal(t, data, newData)
	idFactory.AssertExpectations(t)
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

	//todo should currently fail because not implemented
	assert.Error(t, err)

	// reset to old DID for other test cases
	cfg.Set("identityId", oldDID.ToAddress().String())
}

func TestService_Exists(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, _, eSrv := getServiceWithMockedLayers()
	entity, _ := CreateEntityWithEmbedCD(t, ctxh, did, nil)
	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

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
	entity = InitEntity(t, did, CreateEntityPayload(t, nil))
	v := documents.ValidatorFunc(func(_, _ documents.Document) error {
		return errors.New("validations fail")
	})
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, v)
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	entity, _ = CreateEntityWithEmbedCD(t, ctxh, did, nil)
	err = eSrv.repo.Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)
	idFactory := new(identity.MockFactory)
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil).Once()
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, CreateValidator(idFactory))
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)
	idFactory.AssertExpectations(t)

	// success
	idFactory = new(identity.MockFactory)
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil).Once()
	entity, _ = CreateEntityWithEmbedCD(t, ctxh, did, nil)
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, CreateValidator(idFactory))
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	idFactory.AssertExpectations(t)
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

	mockAnchor := &mockAnchorSrv{}
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
		mockedERSrv, mockAnchor, mockProcessor, nil)

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
	mockAnchor := &mockAnchorSrv{}
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
		mockedERSrv, mockAnchor, mockProcessor, func() documents.ValidatorGroup {
			return documents.ValidatorGroup{}
		})

	//entity relationship is not the latest request therefore request from peer
	model, err := entitySrv.GetEntityByRelationship(ctxh, erID)
	assert.NoError(t, err)
	assert.Equal(t, model.CurrentVersion(), entity.CurrentVersion())
}

func TestService_CreateModel(t *testing.T) {
	payload := documents.CreatePayload{}
	srv := service{}

	// nil  model
	_, _, err := srv.CreateModel(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// empty context
	payload.Data = utils.RandomSlice(32)
	_, _, err = srv.CreateModel(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentConfigAccountID, err))

	// invalid data
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, _, err = srv.CreateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// validator failed
	fact := new(identity.MockFactory)
	srv.factory = fact
	payload.Data = validData(t)
	_, _, err = srv.CreateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))
	fact.AssertExpectations(t)

	// success
	payload.Data = validDataWithIdentity(t)
	srv.repo = testRepo()
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	fact = new(identity.MockFactory)
	fact.On("IdentityExists", mock.Anything).Return(true, nil)
	srv.factory = fact
	m, _, err := srv.CreateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	jm.AssertExpectations(t)
	fact.AssertExpectations(t)
}

func TestService_UpdateModel(t *testing.T) {
	payload := documents.UpdatePayload{}
	srv := service{}

	// nil  model
	_, _, err := srv.UpdateModel(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// empty context
	payload.Data = utils.RandomSlice(32)
	_, _, err = srv.UpdateModel(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentConfigAccountID, err))

	// missing id
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, fact, srvr := getServiceWithMockedLayers()
	srv = srvr.(service)
	payload.DocumentID = utils.RandomSlice(32)
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// payload invalid
	e, _ := CreateEntityWithEmbedCD(t, ctxh, did, nil)
	err = testRepo().Create(did[:], e.ID(), e)
	assert.NoError(t, err)
	payload.DocumentID = e.ID()
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// validator failed
	payload.Data = validData(t)
	srv.factory = fact
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))
	fact.AssertExpectations(t)

	// Success
	fact = new(identity.MockFactory)
	fact.On("IdentityExists", mock.Anything).Return(true, nil)
	srv.factory = fact
	payload.Data = validDataWithIdentity(t)
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	m, _, err := srv.UpdateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.Equal(t, m.ID(), e.ID())
	assert.Equal(t, m.CurrentVersion(), e.NextVersion())
	jm.AssertExpectations(t)
	fact.AssertExpectations(t)
}

func TestService_ValidateError(t *testing.T) {
	srv := service{}
	err := srv.Validate(context.Background(), nil, nil)
	assert.Error(t, err)
}
