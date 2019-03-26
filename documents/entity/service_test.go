// +build unit

package entity

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(dIDBytes, nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)

	repo := testRepo()
	mockAnchor := &mockAnchorRepo{}
	docSrv := documents.DefaultService(repo, mockAnchor, documents.NewServiceRegistry(), &idService)
	return idService, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[transactions.BootstrappedService].(transactions.Manager))
}

func TestService_Update(t *testing.T) {
	_, srv := getServiceWithMockedLayers()
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
	data, err := eSrv.DeriveEntityData(model)
	assert.NoError(t, err)
	data.LegalName = "test company"
	data.Contacts = []*entitypb.Contact{{Name: "Mr. Test"}}
	collab := testingidentity.GenerateRandomDID().String()
	newInv, err := eSrv.DeriveFromUpdatePayload(ctxh, &cliententitypb.EntityUpdatePayload{
		Identifier:    hexutil.Encode(model.ID()),
		Collaborators: []string{collab},
		Data:          data,
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
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
	// nil payload
	doc, err := eSrv.DeriveFromUpdatePayload(nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// nil payload data
	doc, err = eSrv.DeriveFromUpdatePayload(nil, &cliententitypb.EntityUpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
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
	payload.Collaborators = []string{"some wrong ID"}
	payload.Data.Identity = testingidentity.GenerateRandomDID().String()
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPrepareCoreDocument, err))
	assert.Nil(t, doc)

	// success
	wantCollab := testingidentity.GenerateRandomDID()

	payload.Collaborators = []string{wantCollab.String()}
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
	cs, err := doc.GetCollaborators()
	assert.Len(t, cs, 3)
	assert.Contains(t, cs, wantCollab)
	assert.Equal(t, old.ID(), doc.ID())
	assert.Equal(t, payload.Identifier, hexutil.Encode(doc.ID()))
	assert.Equal(t, old.CurrentVersion(), doc.PreviousVersion())
	assert.Equal(t, old.NextVersion(), doc.CurrentVersion())
	assert.NotNil(t, doc.NextVersion())
	assert.Equal(t, payload.Data, doc.(*Entity).getClientData())
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	eSrv := service{}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	m, err := eSrv.DeriveFromCreatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// nil data payload
	m, err = eSrv.DeriveFromCreatePayload(ctxh, &cliententitypb.EntityCreatePayload{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

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
	_, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// calculate data root fails
	m, _, _, err := eSrv.Create(ctxh, &mockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	entity, err := eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.NoError(t, err)
	m, _, _, err = eSrv.Create(ctxh, entity)
	assert.NoError(t, err)
	assert.True(t, testRepo().Exists(accountID, m.ID()))
	assert.True(t, testRepo().Exists(accountID, m.CurrentVersion()))
}

func TestService_DeriveEntityData(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()

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
	// success
	eSrv := service{repo: testRepo()}

	// derive data failed
	m := new(mockModel)
	r, err := eSrv.DeriveEntityResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	entity, _ := createCDWithEmbeddedEntity(t)
	r, err = eSrv.DeriveEntityResponse(entity)
	payload := testingdocuments.CreateEntityPayload()
	assert.NoError(t, err)
	assert.Equal(t, payload.Data.Contacts[0].Name, r.Data.Contacts[0].Name)
	assert.Equal(t, payload.Data.LegalName, r.Data.LegalName)
	assert.Contains(t, r.Header.Collaborators, did.String())
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
	doc, _ := createCDWithEmbeddedEntity(t)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.NoError(t, err)

	data := doc.(*Entity).getClientData()
	data.LegalName = "test company"
	doc2 := new(Entity)
	assert.NoError(t, doc2.PrepareNewVersion(doc, data, nil))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := eSrv.GetCurrentVersion(ctxh, doc.ID())

	doc3Entity := doc3.(*Entity)

	assert.NoError(t, err)
	assert.Equal(t, doc2.LegalName, doc3Entity.LegalName)
}

func TestService_GetVersion(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
	entity, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, entity.CurrentVersion(), entity)
	assert.NoError(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := eSrv.GetVersion(ctxh, entity.ID(), entity.CurrentVersion())
	assert.NoError(t, err)

	mod, err = eSrv.GetVersion(ctxh, mod.ID(), []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
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
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, CreateValidator())
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)

	// success
	entity, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.NoError(t, err)
	entity, err = eSrv.validateAndPersist(ctxh, nil, entity, CreateValidator())
	assert.NoError(t, err)
	assert.NotNil(t, entity)
}
