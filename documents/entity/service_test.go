// +build unit

package entity

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	cid         = testingidentity.GenerateRandomDID()
	centIDBytes = cid[:]
	accountID   = cid[:]
)

type mockAnchorRepo struct {
	mock.Mock
	anchors.AnchorRepository
}

func (r *mockAnchorRepo) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocumentRoot, error) {
	args := r.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocumentRoot)
	return docRoot, args.Error(1)
}

/*
func getServiceWithMockedLayers() (testingcommons.MockIdentityService, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
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
/*
func TestService_Update(t *testing.T) {
	_, srv := getServiceWithMockedLayers()
	invSrv := srv.(service)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// missing last version
	model, _ := createCDWithEmbeddedEntity(t)
	_, _, _, err := invSrv.Update(ctxh, model)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.NoError(t, testRepo().Create(accountID, model.CurrentVersion(), model))

	// calculate data root fails
	nm := new(mockModel)
	nm.On("ID").Return(model.ID(), nil).Once()
	_, _, _, err = invSrv.Update(ctxh, nm)
	nm.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	data, err := invSrv.DeriveEntityData(model)
	assert.Nil(t, err)
	data.GrossAmount = 100
	data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	collab := testingidentity.GenerateRandomDID().String()
	newInv, err := invSrv.DeriveFromUpdatePayload(ctxh, &cliententitypb.EntityUpdatePayload{
		Identifier:    hexutil.Encode(model.ID()),
		Collaborators: []string{collab},
		Data:          data,
	})
	assert.Nil(t, err)
	newData, err := invSrv.DeriveEntityData(newInv)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)

	model, _, _, err = invSrv.Update(ctxh, newInv)
	assert.Nil(t, err)
	assert.NotNil(t, model)
	assert.True(t, testRepo().Exists(accountID, model.ID()))
	assert.True(t, testRepo().Exists(accountID, model.CurrentVersion()))
	assert.True(t, testRepo().Exists(accountID, model.PreviousVersion()))

	newData, err = invSrv.DeriveEntityData(model)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	// nil payload
	doc, err := invSrv.DeriveFromUpdatePayload(nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// nil payload data
	doc, err = invSrv.DeriveFromUpdatePayload(nil, &cliententitypb.EntityUpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// messed up identifier
	contextHeader := testingconfig.CreateAccountContext(t, cfg)
	payload := &cliententitypb.EntityUpdatePayload{Identifier: "some identifier", Data: &cliententitypb.EntityData{}}
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentIdentifier, err))
	assert.Contains(t, err.Error(), "failed to decode identifier")
	assert.Nil(t, doc)

	// missing last version
	id := utils.RandomSlice(32)
	payload.Identifier = hexutil.Encode(id)
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.Nil(t, doc)

	// failed to load from data
	old, _ := createCDWithEmbeddedEntity(t)
	err = testRepo().Create(accountID, old.CurrentVersion(), old)
	assert.Nil(t, err)
	payload.Data = &cliententitypb.EntityData{
		Sender:      "0xed03fa80291ff5ddc284de6b51e716b130b05e20",
		Recipient:   "0xea939d5c0494b072c51565b191ee59b5d34fbf79",
		Payee:       "0x087d8ca6a16e6ce8d9ff55672e551a2828ab8e8c",
		GrossAmount: 42,
		ExtraData:   "some data",
		Currency:    "EUR",
	}

	payload.Identifier = hexutil.Encode(old.ID())
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.Nil(t, doc)

	// failed core document new version
	payload.Data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	payload.Collaborators = []string{"some wrong ID"}
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPrepareCoreDocument, err))
	assert.Nil(t, doc)

	// success
	wantCollab := testingidentity.GenerateRandomDID()

	payload.Collaborators = []string{wantCollab.String()}
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
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

func TestService_DeriveFromCreatePayload(t *testing.T) {
	invSrv := service{}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	m, err := invSrv.DeriveFromCreatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// nil data payload
	m, err = invSrv.DeriveFromCreatePayload(ctxh, &cliententitypb.EntityCreatePayload{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// Init fails
	payload := &cliententitypb.EntityCreatePayload{
		Data: &cliententitypb.EntityData{
			ExtraData: "some data",
		},
	}

	m, err = invSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.Data.ExtraData = "0x01020304050607"
	m, err = invSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	inv := m.(*Entity)
	assert.Equal(t, hexutil.Encode(inv.ExtraData), payload.Data.ExtraData)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	invSrv := service{repo: testRepo()}
	_, cd := createCDWithEmbeddedEntity(t)
	m, err := invSrv.DeriveFromCoreDocument(cd)
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
	invSrv := srv.(service)

	// calculate data root fails
	m, _, _, err := invSrv.Create(ctxh, &mockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	inv, err := invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	m, _, _, err = invSrv.Create(ctxh, inv)
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(accountID, m.ID()))
	assert.True(t, testRepo().Exists(accountID, m.CurrentVersion()))
}

func TestService_DeriveEntityData(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()

	// some random model
	_, err := invSrv.DeriveEntityData(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := testingdocuments.CreateEntityPayload()
	inv, err := invSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err, "must be non nil")
	data, err := invSrv.DeriveEntityData(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
}

func TestService_DeriveEntityResponse(t *testing.T) {
	// success
	invSrv := service{repo: testRepo()}

	// derive data failed
	m := new(mockModel)
	r, err := invSrv.DeriveEntityResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	inv, _ := createCDWithEmbeddedEntity(t)
	r, err = invSrv.DeriveEntityResponse(inv)
	payload := testingdocuments.CreateEntityPayload()
	assert.Nil(t, err)
	assert.Equal(t, payload.Data, r.Data)
	assert.Contains(t, r.Header.Collaborators, cid.String())
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	doc, _ := createCDWithEmbeddedEntity(t)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.Nil(t, err)

	data := doc.(*Entity).getClientData()
	data.Currency = "INR"
	doc2 := new(Entity)
	assert.NoError(t, doc2.PrepareNewVersion(doc, data, nil))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := invSrv.GetCurrentVersion(ctxh, doc.ID())
	assert.Nil(t, err)
	assert.Equal(t, doc2, doc3)
}

func TestService_GetVersion(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	inv, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := invSrv.GetVersion(ctxh, inv.ID(), inv.CurrentVersion())
	assert.Nil(t, err)

	mod, err = invSrv.GetVersion(ctxh, mod.ID(), []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	inv, _ := createCDWithEmbeddedEntity(t)
	err := testRepo().Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := invSrv.Exists(ctxh, inv.CurrentVersion())
	assert.True(t, exists, "entity should exist")

	exists = invSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, " entity should not exist")
}

func TestService_calculateDataRoot(t *testing.T) {
	invSrv := service{repo: testRepo()}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// type mismatch
	inv, err := invSrv.validateAndPersist(ctxh, nil, &mockModel{}, nil)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	inv, err = invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, v)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	inv, err = invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	err = invSrv.repo.Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)

	// success
	inv, err = invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateEntityPayload())
	assert.Nil(t, err)
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
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
func createCDWithEmbeddedEntity(t *testing.T) (documents.Model, coredocumentpb.CoreDocument) {
	i := new(Entity)
	err := i.InitEntityInput(testingdocuments.CreateEntityPayload(), cid.String())
	assert.NoError(t, err)
	_, err = i.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := i.PackCoreDocument()
	assert.NoError(t, err)
	return i, cd
}
