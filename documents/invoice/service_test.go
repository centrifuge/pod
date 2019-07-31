// +build unit

package invoice

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/anchors"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	did       = testingidentity.GenerateRandomDID()
	didBytes  = did[:]
	accountID = did[:]
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, documents.Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(didBytes, nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)

	repo := testRepo()
	anchorRepo := &testinganchors.MockAnchorRepo{}
	anchorRepo.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(cfg, repo, anchorRepo, documents.NewServiceRegistry(), &idService, nil, nil)
	return idService, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[jobs.BootstrappedService].(jobs.Manager),
		func() documents.TokenRegistry { return nil }, anchorRepo)
}

func TestService_Update(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// missing last version
	inv, _ := CreateInvoiceWithEmbedCD(t, ctxh, did, nil)
	_, _, _, err := invSrv.Update(ctxh, inv)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// success
	assert.NoError(t, testRepo().Create(accountID, inv.CurrentVersion(), inv))
	dec, err := documents.NewDecimal("100")
	assert.NoError(t, err)
	data := inv.GetData().(Data)
	data.NetAmount = dec
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	collab := testingidentity.GenerateRandomDID()
	newInv := new(Invoice)
	assert.NoError(t, newInv.unpackFromUpdatePayloadOld(inv, documents.UpdatePayload{
		DocumentID: inv.ID(),
		CreatePayload: documents.CreatePayload{
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []identity.DID{collab},
			},
			Data: d,
		},
	}))
	newData := newInv.GetData()
	assert.Equal(t, data, newData)
	invModel, _, _, err := invSrv.Update(ctxh, newInv)
	assert.Nil(t, err)
	assert.NotNil(t, invModel)
	assert.True(t, testRepo().Exists(accountID, invModel.ID()))
	assert.True(t, testRepo().Exists(accountID, invModel.CurrentVersion()))
	assert.True(t, testRepo().Exists(accountID, invModel.PreviousVersion()))

	newData = invModel.GetData()
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	invSrv := service{repo: testRepo()}
	_, cd := createCDWithEmbeddedInvoice(t)
	m, err := invSrv.DeriveFromCoreDocument(cd)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	inv, ok := m.(*Invoice)
	assert.True(t, ok, "must be true")
	assert.Equal(t, inv.Data.Recipient.String(), "0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.Equal(t, inv.Data.Currency, "EUR")
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
	inv, _ := CreateInvoiceWithEmbedCDWithPayload(t, ctxh, did, CreateInvoicePayload(t, nil))
	m, _, _, err = invSrv.Create(ctxh, inv)
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(accountID, m.ID()))
	assert.True(t, testRepo().Exists(accountID, m.CurrentVersion()))
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc, _ := CreateInvoiceWithEmbedCD(t, ctxh, did, nil)
	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.Nil(t, err)

	data := doc.Data
	data.Currency = "INR"
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	doc2 := new(Invoice)
	assert.NoError(t, doc2.unpackFromUpdatePayloadOld(doc, documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{Data: d},
		DocumentID:    doc.ID(),
	}))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := invSrv.GetCurrentVersion(ctxh, doc.ID())
	assert.Nil(t, err)
	assert.Equal(t, doc2, doc3)
}

func TestService_GetVersion(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	inv, _ := createCDWithEmbeddedInvoice(t)
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
	inv, _ := createCDWithEmbeddedInvoice(t)
	err := testRepo().Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := invSrv.Exists(ctxh, inv.CurrentVersion())
	assert.True(t, exists, "invoice should exist")

	exists = invSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, " invoice should not exist")
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
	inv = new(Invoice)
	assert.NoError(t, inv.(*Invoice).unpackFromCreatePayload(did, CreateInvoicePayload(t, nil)))
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, v)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	inv = new(Invoice)
	assert.NoError(t, inv.(*Invoice).unpackFromCreatePayload(did, CreateInvoicePayload(t, nil)))
	err = invSrv.repo.Create(accountID, inv.CurrentVersion(), inv)
	assert.Nil(t, err)
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)

	// success
	inv = new(Invoice)
	assert.NoError(t, inv.(*Invoice).unpackFromCreatePayload(did, CreateInvoicePayload(t, nil)))
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
	assert.Nil(t, err)
	assert.NotNil(t, inv)
}

var testRepoGlobal documents.Repository

func testRepo() documents.Repository {
	if testRepoGlobal != nil {
		return testRepoGlobal
	}

	ldb, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
	if err != nil {
		panic(err)
	}
	testRepoGlobal = documents.NewDBRepository(leveldb.NewLevelDBRepository(ldb))
	testRepoGlobal.Register(&Invoice{})
	return testRepoGlobal
}

func createCDWithEmbeddedInvoice(t *testing.T) (documents.Model, coredocumentpb.CoreDocument) {
	i := new(Invoice)
	i = InitInvoice(t, did, CreateInvoicePayload(t, nil))
	i.GetTestCoreDocWithReset()
	_, err := i.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := i.PackCoreDocument()
	assert.NoError(t, err)
	return i, cd
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
	payload.Data = validData(t)
	_, _, err = srv.CreateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.Data = validDataWithCurrency(t)
	srv.repo = testRepo()
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	m, _, err := srv.CreateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	jm.AssertExpectations(t)
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
	_, srvr := getServiceWithMockedLayers()
	srv = srvr.(service)
	payload.DocumentID = utils.RandomSlice(32)
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// payload invalid
	inv, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	err = testRepo().Create(did[:], inv.ID(), inv)
	assert.NoError(t, err)
	payload.DocumentID = inv.ID()
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// validator failed
	payload.Data = validData(t)
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// Success
	payload.Data = validDataWithCurrency(t)
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	m, _, err := srv.UpdateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.Equal(t, m.ID(), inv.ID())
	assert.Equal(t, m.CurrentVersion(), inv.NextVersion())
	jm.AssertExpectations(t)
}

func TestService_Derive(t *testing.T) {
	// new document
	payload := documents.UpdatePayload{CreatePayload: documents.CreatePayload{Scheme: Scheme, Data: invalidDecimalData(t)}}
	s := service{}

	// missing account ctx
	ctx := context.Background()
	_, err := s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentConfigAccountID, err))

	// invalid payload
	ctx = testingconfig.CreateAccountContext(t, cfg)
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// valid data
	payload.Data = validData(t)
	inv, err := s.Derive(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, inv)

	// update document
	docID := utils.RandomSlice(32)
	payload.DocumentID = docID
	srv := new(testingdocuments.MockService)
	srv.On("GetCurrentVersion", docID).Return(nil, documents.ErrDocumentNotFound).Once()
	s.Service = srv
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// invalid type
	srv.On("GetCurrentVersion", docID).Return(new(mockModel), nil).Once()
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// invalid data
	payload.Data = invalidDIDData(t)
	old, _ := CreateInvoiceWithEmbedCD(t, ctx, did, nil)
	srv.On("GetCurrentVersion", docID).Return(old, nil)
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.Data = validData(t)
	inv, err = s.Derive(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, inv)
	srv.AssertExpectations(t)
}

func TestService_Validate(t *testing.T) {
	srv := service{}
	err := srv.Validate(context.Background(), nil)
	assert.NoError(t, err)
}
