// +build unit

package purchaseorder

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/anchors"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	did       = testingidentity.GenerateRandomDID()
	accountID = did[:]
)

func getServiceWithMockedLayers() (*testingcommons.MockIdentityService, documents.Service) {
	idService := &testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	jobManager := ctx[jobs.BootstrappedService].(jobs.Manager)
	repo := testRepo()
	anchorRepo := &testinganchors.MockAnchorRepo{}
	anchorRepo.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(cfg, repo, anchorRepo, documents.NewServiceRegistry(), idService)
	return idService, DefaultService(docSrv, repo, queueSrv, jobManager, anchorRepo)
}

func TestService_Update(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// missing last version
	po, _ := CreatePOWithEmbedCD(t, ctxh, did, nil)
	_, _, _, err := poSrv.Update(ctxh, po)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// success
	assert.NoError(t, testRepo().Create(accountID, po.CurrentVersion(), po))
	dec, err := documents.NewDecimal("100")
	assert.NoError(t, err)
	data := po.GetData().(Data)
	data.TotalAmount = dec
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	collab := testingidentity.GenerateRandomDID()
	newPO := new(PurchaseOrder)
	assert.NoError(t, newPO.unpackFromUpdatePayload(po, documents.UpdatePayload{
		DocumentID: po.ID(),
		CreatePayload: documents.CreatePayload{
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []identity.DID{collab},
			},
			Data: d,
		},
	}))
	newData := newPO.GetData()
	assert.Equal(t, data, newData)
	poModel, _, _, err := poSrv.Update(ctxh, newPO)
	assert.Nil(t, err)
	assert.NotNil(t, poModel)
	assert.True(t, testRepo().Exists(accountID, poModel.ID()))
	assert.True(t, testRepo().Exists(accountID, poModel.CurrentVersion()))
	assert.True(t, testRepo().Exists(accountID, poModel.PreviousVersion()))

	newData = poModel.GetData()
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	poSrv := service{repo: testRepo()}
	_, cd := CreatePOWithEmbedCD(t, nil, did, nil)
	m, err := poSrv.DeriveFromCoreDocument(cd)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	po, ok := m.(*PurchaseOrder)
	assert.True(t, ok, "must be true")
	assert.Equal(t, po.Data.Recipient.String(), "0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.Equal(t, po.Data.TotalAmount.String(), "42")
}

func TestService_GetCurrentVersion(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc, _ := CreatePOWithEmbedCD(t, ctxh, did, nil)
	err := testRepo().Create(accountID, doc.CurrentVersion(), doc)
	assert.Nil(t, err)

	data := doc.Data
	data.Currency = "INR"
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	doc2 := new(PurchaseOrder)
	assert.NoError(t, doc2.unpackFromUpdatePayload(doc, documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{Data: d},
		DocumentID:    doc.ID(),
	}))
	assert.NoError(t, testRepo().Create(accountID, doc2.CurrentVersion(), doc2))

	doc3, err := poSrv.GetCurrentVersion(ctxh, doc.ID())
	assert.Nil(t, err)
	assert.Equal(t, doc2, doc3)
}

func TestService_GetVersion(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	err := testRepo().Create(accountID, po.CurrentVersion(), po)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := poSrv.GetVersion(ctxh, po.ID(), po.CurrentVersion())
	assert.Nil(t, err)

	mod, err = poSrv.GetVersion(ctxh, mod.ID(), []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	err := testRepo().Create(accountID, po.CurrentVersion(), po)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := poSrv.Exists(ctxh, po.CurrentVersion())
	assert.True(t, exists, "purchase order should exist")

	exists = poSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, "purchase order should not exist")
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
	testRepoGlobal.Register(&PurchaseOrder{})
	return testRepoGlobal
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
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	err = testRepo().Create(did[:], po.ID(), po)
	assert.NoError(t, err)
	payload.DocumentID = po.ID()
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
	assert.Equal(t, m.ID(), po.ID())
	assert.Equal(t, m.CurrentVersion(), po.NextVersion())
	jm.AssertExpectations(t)
}
