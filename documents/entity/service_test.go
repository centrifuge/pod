// +build unit

package entity

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"

	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
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
	assert.Nil(t, err)
	data.LegalName = "test company"
	data.Contacts = []*entitypb.Contact{{Name:"Mr. Test"}}
	collab := testingidentity.GenerateRandomDID().String()
	newInv, err := eSrv.DeriveFromUpdatePayload(ctxh, &cliententitypb.EntityUpdatePayload{
		Identifier:    hexutil.Encode(model.ID()),
		Collaborators: []string{collab},
		Data:          data,
	})
	assert.Nil(t, err)
	newData, err := eSrv.DeriveEntityData(newInv)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)

	model, _, _, err = eSrv.Update(ctxh, newInv)
	assert.Nil(t, err)
	assert.NotNil(t, model)
	assert.True(t, testRepo().Exists(accountID, model.ID()))
	assert.True(t, testRepo().Exists(accountID, model.CurrentVersion()))
	assert.True(t, testRepo().Exists(accountID, model.PreviousVersion()))

	newData, err = eSrv.DeriveEntityData(model)
	assert.Nil(t, err)
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

	// failed to load from data
	old, _ := createCDWithEmbeddedEntity(t)
	err = testRepo().Create(accountID, old.CurrentVersion(), old)
	assert.Nil(t, err)
	payload.Data = &cliententitypb.EntityData{
		LegalName: "test company",
		Contacts: []*entitypb.Contact{{Name:"Mr. Test"}},
	}

	payload.Identifier = hexutil.Encode(old.ID())
	doc, err = eSrv.DeriveFromUpdatePayload(contextHeader, payload)

	assert.Error(t, err, "should fail because Identity is missing")
	assert.Nil(t, doc)

	// failed core document new version
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
	assert.Nil(t, err)
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
/*
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
			ExtraData: "some data",
		},
	}

	m, err = eSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.Data.ExtraData = "0x01020304050607"
	m, err = eSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	inv := m.(*Entity)
	assert.Equal(t, hexutil.Encode(inv.ExtraData), payload.Data.ExtraData)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	eSrv := service{repo: testRepo()}
	_, cd := createCDWithEmbeddedEntity(t)
	m, err := eSrv.DeriveFromCoreDocument(cd)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	inv, ok := m.(*Entity)
	assert.True(t, ok, "must be true")
	assert.Equal(t, inv.Recipient.String(), "0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.Equal(t, inv.GrossAmount, int64(42))
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
	inv, err := eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	m, _, _, err = eSrv.Create(ctxh, inv)
	assert.Nil(t, err)
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
	inv, err := eSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err, "must be non nil")
	data, err := eSrv.DeriveEntityData(inv)
	assert.Nil(t, err, "Derive must succeed")
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
	inv, _ := createCDWithEmbeddedEntity(t)
	r, err = eSrv.DeriveEntityResponse(inv)
	payload := testingdocuments.CreateEntityPayload()
	assert.Nil(t, err)
	assert.Equal(t, payload.Data, r.Data)
	assert.Contains(t, r.Header.Collaborators, cid.String())
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
	doc, _ := createCDWithEmbeddedEntity(t)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.Nil(t, err)

	data := doc.(*Entity).getClientData()
	data.Currency = "INR"
	doc2 := new(Entity)
	assert.NoError(t, doc2.PrepareNewVersion(doc, data, nil))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := eSrv.GetCurrentVersion(ctxh, doc.ID())
	assert.Nil(t, err)
	assert.Equal(t, doc2, doc3)
}

func TestService_GetVersion(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
	inv, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := eSrv.GetVersion(ctxh, inv.ID(), inv.CurrentVersion())
	assert.Nil(t, err)

	mod, err = eSrv.GetVersion(ctxh, mod.ID(), []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, eSrv := getServiceWithMockedLayers()
	inv, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := eSrv.Exists(ctxh, inv.CurrentVersion())
	assert.True(t, exists, "entity should exist")

	exists = eSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, " entity should not exist")
}

func TestService_calculateDataRoot(t *testing.T) {
	eSrv := service{repo: testRepo()}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// type mismatch
	inv, err := eSrv.validateAndPersist(ctxh, nil, &mockModel{}, nil)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	inv, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	inv, err = eSrv.validateAndPersist(ctxh, nil, inv, v)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	inv, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	err = eSrv.repo.Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)
	inv, err = eSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)

	// success
	inv, err = eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	inv, err = eSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
	assert.Nil(t, err)
	assert.NotNil(t, inv)
}

var testRepoGlobal documents.Repository

func testRepo() documents.Repository {
	if testRepoGlobal == nil {
		ldb, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
		if err != nil {
			panic(err)
		}
		testRepoGlobal = documents.NewDBRepository(leveldb.NewLevelDBRepository(ldb))
		testRepoGlobal.Register(&Entity{})
	}
	return testRepoGlobal
}
*/

