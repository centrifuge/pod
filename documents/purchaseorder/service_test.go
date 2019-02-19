// +build unit

package purchaseorder

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	cid         = identity.RandomCentID()
	accountID   = cid[:]
	centIDBytes = cid[:]
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

func getServiceWithMockedLayers() (*testingcommons.MockIdentityService, Service) {
	idService := &testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	txManager := ctx[transactions.BootstrappedService].(transactions.Manager)
	repo := testRepo()
	mockAnchor := &mockAnchorRepo{}
	docSrv := documents.DefaultService(repo, mockAnchor, documents.NewServiceRegistry(), idService)
	return idService, DefaultService(docSrv, repo, queueSrv, txManager)
}

func TestService_Update(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// pack failed
	model := &testingdocuments.MockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("pack error")).Once()
	_, _, _, err := poSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack error")

	// missing last version
	model = &testingdocuments.MockModel{}
	dm := documents.NewCoreDocModel()
	model.On("PackCoreDocument").Return(dm, nil).Once()
	_, _, _, err = poSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")

	payload := testingdocuments.CreatePOPayload()
	payload.Collaborators = []string{"0x010203040506"}
	po, err := poSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	dm, err = po.PackCoreDocument()
	assert.Nil(t, err)
	dm.Document.DocumentRoot = utils.RandomSlice(32)
	po.(*PurchaseOrder).CoreDocumentModel = dm
	testRepo().Create(accountID, dm.Document.CurrentVersion, po)

	// calculate data root fails
	model = &testingdocuments.MockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Once()
	_, _, _, err = poSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	data, err := poSrv.DerivePurchaseOrderData(po)
	assert.Nil(t, err)
	data.OrderAmount = 100
	data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	collab := hexutil.Encode(utils.RandomSlice(6))
	newPO, err := poSrv.DeriveFromUpdatePayload(ctxh, &clientpurchaseorderpb.PurchaseOrderUpdatePayload{
		Identifier:    hexutil.Encode(dm.Document.DocumentIdentifier),
		Collaborators: []string{collab},
		Data:          data,
	})
	assert.Nil(t, err)
	newData, err := poSrv.DerivePurchaseOrderData(newPO)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
	po, _, _, err = poSrv.Update(ctxh, newPO)
	assert.Nil(t, err)
	assert.NotNil(t, po)

	newDM, err := po.PackCoreDocument()
	newCD := newDM.Document
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(accountID, newCD.DocumentIdentifier))
	assert.True(t, testRepo().Exists(accountID, newCD.CurrentVersion))
	assert.True(t, testRepo().Exists(accountID, newCD.PreviousVersion))

	newData, err = poSrv.DerivePurchaseOrderData(po)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	doc, err := poSrv.DeriveFromUpdatePayload(ctxh, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// nil payload data
	doc, err = poSrv.DeriveFromUpdatePayload(ctxh, &clientpurchaseorderpb.PurchaseOrderUpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// messed up identifier
	contextHeader := testingconfig.CreateAccountContext(t, cfg)
	payload := &clientpurchaseorderpb.PurchaseOrderUpdatePayload{Identifier: "some identifier", Data: &clientpurchaseorderpb.PurchaseOrderData{}}
	doc, err = poSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode identifier")
	assert.Nil(t, doc)

	// missing last version
	id := utils.RandomSlice(32)
	payload.Identifier = hexutil.Encode(id)
	doc, err = poSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.Nil(t, doc)

	// failed to load from data
	self, _ := contextutil.Self(contextHeader)
	old := new(PurchaseOrder)
	err = old.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), self.ID.String())
	assert.Nil(t, err)
	oldCD := old.CoreDocumentModel.Document
	oldCD.DocumentIdentifier = id
	oldCD.CurrentVersion = id
	oldCD.DocumentRoot = utils.RandomSlice(32)
	err = testRepo().Create(accountID, id, old)
	assert.Nil(t, err)
	payload.Data = &clientpurchaseorderpb.PurchaseOrderData{
		Recipient: "0x010203040506",
		ExtraData: "some data",
		Currency:  "EUR",
	}

	doc, err = poSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load purchase order from data")
	assert.Nil(t, doc)

	// failed core document new version
	payload.Data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	payload.Collaborators = []string{"some wrong ID"}
	doc, err = poSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPrepareCoreDocument, err))
	assert.Nil(t, doc)

	// success
	wantCollab := utils.RandomSlice(6)
	payload.Collaborators = []string{hexutil.Encode(wantCollab)}
	doc, err = poSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Nil(t, err)
	assert.NotNil(t, doc)
	dm, err := doc.PackCoreDocument()
	cd := dm.Document
	assert.Nil(t, err)
	assert.Equal(t, wantCollab, cd.Collaborators[2])
	assert.Len(t, cd.Collaborators, 3)
	oldDM, err := old.PackCoreDocument()
	oldCD = oldDM.Document
	assert.Nil(t, err)
	assert.Equal(t, oldCD.DocumentIdentifier, cd.DocumentIdentifier)
	assert.Equal(t, payload.Identifier, hexutil.Encode(cd.DocumentIdentifier))
	assert.Equal(t, oldCD.CurrentVersion, cd.PreviousVersion)
	assert.Equal(t, oldCD.NextVersion, cd.CurrentVersion)
	assert.NotNil(t, cd.NextVersion)
	assert.Equal(t, payload.Data, doc.(*PurchaseOrder).getClientData())
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	poSrv := service{}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	m, err := poSrv.DeriveFromCreatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// nil data payload
	m, err = poSrv.DeriveFromCreatePayload(ctxh, &clientpurchaseorderpb.PurchaseOrderCreatePayload{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

	// Init fails
	payload := &clientpurchaseorderpb.PurchaseOrderCreatePayload{
		Data: &clientpurchaseorderpb.PurchaseOrderData{
			ExtraData: "some data",
		},
	}

	m, err = poSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.Data.ExtraData = "0x01020304050607"
	m, err = poSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	po := m.(*PurchaseOrder)
	assert.Equal(t, hexutil.Encode(po.ExtraData), payload.Data.ExtraData)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	// nil doc
	poSrv := service{repo: testRepo()}
	_, err := poSrv.DeriveFromCoreDocumentModel(nil)
	assert.Error(t, err, "must fail to derive")

	// successful
	data := testingdocuments.CreatePOData()
	dm := CreateCDWithEmbeddedPO(t, data)
	m, err := poSrv.DeriveFromCoreDocumentModel(dm)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	po, ok := m.(*PurchaseOrder)
	assert.True(t, ok, "must be true")
	assert.Equal(t, po.Recipient[:], data.Recipient)
	assert.Equal(t, po.OrderAmount, data.OrderAmount)
}

func TestService_Create(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()

	// calculate data root fails
	m, _, _, err := poSrv.Create(ctxh, &testingdocuments.MockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// anchor fails
	po, err := poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	m, _, _, err = poSrv.Create(ctxh, po)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	newDM, err := m.PackCoreDocument()
	newCD := newDM.Document
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(accountID, newCD.DocumentIdentifier))
	assert.True(t, testRepo().Exists(accountID, newCD.CurrentVersion))
}

func TestService_DerivePurchaseOrderData(t *testing.T) {
	var m documents.Model
	_, poSrv := getServiceWithMockedLayers()

	// unknown type
	m = &testingdocuments.MockModel{}
	d, err := poSrv.DerivePurchaseOrderData(m)
	assert.Nil(t, d)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	payload := testingdocuments.CreatePOPayload()
	m, err = poSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err)
	d, err = poSrv.DerivePurchaseOrderData(m)
	assert.Nil(t, err)
	assert.Equal(t, d.Currency, payload.Data.Currency)
}

func TestService_DerivePurchaseOrderResponse(t *testing.T) {
	poSrv := service{}

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, errors.New("pack core document failed")).Once()
	r, err := poSrv.DerivePurchaseOrderResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack core document failed")

	// cent id failed
	dm := documents.NewCoreDocModel()
	cd := dm.Document
	cd.Collaborators = [][]byte{{1, 2, 3, 4, 5, 6}, {5, 6, 7}}
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(dm, nil).Once()
	r, err = poSrv.DerivePurchaseOrderResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length byte slice provided for centID")

	// derive data failed
	cd.Collaborators = [][]byte{{1, 2, 3, 4, 5, 6}}
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(dm, nil).Once()
	r, err = poSrv.DerivePurchaseOrderResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	payload := testingdocuments.CreatePOPayload()
	po, err := poSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err)
	r, err = poSrv.DerivePurchaseOrderResponse(po)
	assert.Nil(t, err)
	assert.Equal(t, payload.Data, r.Data)
	assert.Equal(t, []string{cid.String(), "0x010101010101"}, r.Header.Collaborators)
}

func createMockDocument() (*PurchaseOrder, error) {
	documentIdentifier := utils.RandomSlice(32)
	nextIdentifier := utils.RandomSlice(32)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: documentIdentifier,
		CurrentVersion:     documentIdentifier,
		NextVersion:        nextIdentifier,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	model := &PurchaseOrder{
		PoNumber:          "test_po",
		OrderAmount:       42,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, documentIdentifier, model)
	return model, err
}

func TestService_GetVersion_wrongTyp(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	currentVersion := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: documentIdentifier,
		CurrentVersion:     currentVersion,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	//should be an po
	po := &invoice.Invoice{
		GrossAmount:       60,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, currentVersion, po)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, err = poSrv.GetVersion(ctxh, documentIdentifier, currentVersion)
	assert.Error(t, err)

}

func TestService_GetCurrentVersion(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()
	thirdIdentifier := utils.RandomSlice(32)
	doc, err := createMockDocument()
	assert.Nil(t, err)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	mod1, err := poSrv.GetCurrentVersion(ctxh, doc.CoreDocumentModel.Document.DocumentIdentifier)
	assert.Nil(t, err)

	poLoad1, _ := mod1.(*PurchaseOrder)
	assert.Equal(t, poLoad1.CoreDocumentModel.Document.CurrentVersion, doc.CoreDocumentModel.Document.DocumentIdentifier)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: doc.CoreDocumentModel.Document.DocumentIdentifier,
		CurrentVersion:     doc.CoreDocumentModel.Document.NextVersion,
		NextVersion:        thirdIdentifier,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	po2 := &PurchaseOrder{
		OrderAmount:       42,
		CoreDocumentModel: coreDocModel,
	}

	err = testRepo().Create(accountID, doc.CoreDocumentModel.Document.NextVersion, po2)
	assert.Nil(t, err)

	mod2, err := poSrv.GetCurrentVersion(ctxh, doc.CoreDocumentModel.Document.DocumentIdentifier)
	assert.Nil(t, err)

	poLoad2, _ := mod2.(*PurchaseOrder)
	assert.Equal(t, poLoad2.CoreDocumentModel.Document.CurrentVersion, doc.CoreDocumentModel.Document.NextVersion)
	assert.Equal(t, poLoad2.CoreDocumentModel.Document.NextVersion, thirdIdentifier)
}

func TestService_GetVersion(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: documentIdentifier,
		CurrentVersion:     currentVersion,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}

	po := &PurchaseOrder{
		OrderAmount:       42,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, currentVersion, po)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := poSrv.GetVersion(ctxh, documentIdentifier, currentVersion)
	assert.Nil(t, err)
	loadpo, _ := mod.(*PurchaseOrder)
	assert.Equal(t, loadpo.CoreDocumentModel.Document.CurrentVersion, currentVersion)
	assert.Equal(t, loadpo.CoreDocumentModel.Document.DocumentIdentifier, documentIdentifier)

	mod, err = poSrv.GetVersion(ctxh, documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: documentIdentifier,
		CurrentVersion:     documentIdentifier,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	po := &PurchaseOrder{
		OrderAmount:       42,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, documentIdentifier, po)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := poSrv.Exists(ctxh, documentIdentifier)
	assert.True(t, exists, "purchase order should exist")

	exists = poSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, "purchase order should not exist")

}

func TestService_calculateDataRoot(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	poSrv := service{repo: testRepo()}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// type mismatch
	po, err := poSrv.validateAndPersist(ctxh, nil, &testingdocuments.MockModel{}, nil)
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	po, err = poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	po, err = poSrv.validateAndPersist(ctxh, nil, po, v)
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	po, err = poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrder).CoreDocumentModel.Document.DataRoot)
	err = poSrv.repo.Create(accountID, po.(*PurchaseOrder).CoreDocumentModel.Document.CurrentVersion, po)
	assert.Nil(t, err)
	po, err = poSrv.validateAndPersist(ctxh, nil, po, CreateValidator())
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)

	// success
	po, err = poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrder).CoreDocumentModel.Document.DataRoot)
	po, err = poSrv.validateAndPersist(ctxh, nil, po, CreateValidator())
	assert.Nil(t, err)
	assert.NotNil(t, po)
}

var testRepoGlobal documents.Repository

func testRepo() documents.Repository {
	if testRepoGlobal == nil {
		ldb, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
		if err != nil {
			panic(err)
		}
		testRepoGlobal = documents.NewDBRepository(leveldb.NewLevelDBRepository(ldb))
		testRepoGlobal.Register(&PurchaseOrder{})
	}
	return testRepoGlobal
}
