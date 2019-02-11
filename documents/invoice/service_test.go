// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
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

func TestDefaultService(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil).Once()
	srv := DefaultService(nil, testRepo(), nil, nil)
	assert.NotNil(t, srv, "must be non-nil")
}

func getServiceWithMockedLayers() (testingcommons.MockIDService, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	idService := testingcommons.MockIDService{}
	idService.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil)
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)

	repo := testRepo()
	mockAnchor := &mockAnchorRepo{}
	docSrv := documents.DefaultService(repo, &idService, mockAnchor, documents.NewServiceRegistry())
	return idService, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[transactions.BootstrappedService].(transactions.Manager))
}

func createMockDocument() (*Invoice, error) {
	documentIdentifier := utils.RandomSlice(32)
	nextIdentifier := utils.RandomSlice(32)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: documentIdentifier,
		CurrentVersion:     documentIdentifier,
		NextVersion:        nextIdentifier,
	}
	dm := documents.NewCoreDocModel()
	dm.Document = coreDoc
	inv1 := &Invoice{
		InvoiceNumber:     "test_invoice",
		GrossAmount:       60,
		CoreDocumentModel: dm,
	}
	err := testRepo().Create(accountID, documentIdentifier, inv1)
	return inv1, err
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	// nil doc
	invSrv := service{repo: testRepo()}
	_, err := invSrv.DeriveFromCoreDocumentModel(nil)
	assert.Error(t, err, "must fail to derive")

	// successful
	data := testingdocuments.CreateInvoiceData()
	coreDocModel := testingdocuments.CreateCDWithEmbeddedInvoice(t, data)
	model, err := invSrv.DeriveFromCoreDocumentModel(coreDocModel)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, model, "model must be non-nil")
	inv, ok := model.(*Invoice)
	assert.True(t, ok, "must be true")
	assert.Equal(t, inv.Payee[:], data.Payee)
	assert.Equal(t, inv.Sender[:], data.Sender)
	assert.Equal(t, inv.Recipient[:], data.Recipient)
	assert.Equal(t, inv.GrossAmount, data.GrossAmount)
}

func TestService_DeriveFromPayload(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	payload := testingdocuments.CreateInvoicePayload()
	var model documents.Model
	var err error

	// fail due to nil payload
	_, err = invSrv.DeriveFromCreatePayload(nil, nil)
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")

	// fail due to nil payload data
	_, err = invSrv.DeriveFromCreatePayload(nil, &clientinvoicepb.InvoiceCreatePayload{})
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")

	model, err = invSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err, "valid invoiceData shouldn't produce an error")

	receivedCoreDocumentModel, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")
	assert.NotNil(t, receivedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be field")
}

func TestService_GetLastVersion(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	thirdIdentifier := utils.RandomSlice(32)
	doc, err := createMockDocument()
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod1, err := invSrv.GetCurrentVersion(ctxh, doc.CoreDocumentModel.Document.DocumentIdentifier)
	assert.Nil(t, err)

	invLoad1, _ := mod1.(*Invoice)
	assert.Equal(t, invLoad1.CoreDocumentModel.Document.CurrentVersion, doc.CoreDocumentModel.Document.DocumentIdentifier)

	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: doc.CoreDocumentModel.Document.DocumentIdentifier,
		CurrentVersion:     doc.CoreDocumentModel.Document.NextVersion,
		NextVersion:        thirdIdentifier,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	inv2 := &Invoice{
		GrossAmount:       60,
		CoreDocumentModel: coreDocModel,
	}
	cd := doc.CoreDocumentModel.Document
	err = testRepo().Create(accountID, cd.NextVersion, inv2)
	assert.Nil(t, err)

	mod2, err := invSrv.GetCurrentVersion(ctxh, cd.DocumentIdentifier)
	assert.Nil(t, err)

	invLoad2, _ := mod2.(*Invoice)
	assert.Equal(t, invLoad2.CoreDocumentModel.Document.CurrentVersion, cd.NextVersion)
	assert.Equal(t, invLoad2.CoreDocumentModel.Document.NextVersion, thirdIdentifier)
}

func TestService_GetVersion_wrongTyp(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
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
	//should be an invoice
	po := &purchaseorder.PurchaseOrder{
		NetAmount:         60,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, currentVersion, po)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, err = invSrv.GetVersion(ctxh, documentIdentifier, currentVersion)
	assert.Error(t, err)

}

func TestService_GetVersion(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
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
	inv := &Invoice{
		GrossAmount:       60,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, currentVersion, inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	mod, err := invSrv.GetVersion(ctxh, documentIdentifier, currentVersion)
	assert.Nil(t, err)
	loadInv, _ := mod.(*Invoice)
	assert.Equal(t, loadInv.CoreDocumentModel.Document.CurrentVersion, currentVersion)
	assert.Equal(t, loadInv.CoreDocumentModel.Document.DocumentIdentifier, documentIdentifier)

	mod, err = invSrv.GetVersion(ctxh, documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: documentIdentifier,
		CurrentVersion:     documentIdentifier,
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	inv := &Invoice{
		GrossAmount:       60,
		CoreDocumentModel: coreDocModel,
	}
	err := testRepo().Create(accountID, documentIdentifier, inv)
	assert.Nil(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	exists := invSrv.Exists(ctxh, documentIdentifier)
	assert.True(t, exists, "invoice should exist")

	exists = invSrv.Exists(ctxh, utils.RandomSlice(32))
	assert.False(t, exists, "invoice should not exist")
}

func TestService_Create(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, srv := getServiceWithMockedLayers()
	invSrv := srv.(service)

	// calculate data root fails
	m, _, _, err := invSrv.Create(ctxh, &testingdocuments.MockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// anchor success
	po, err := invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateInvoicePayload())
	assert.Nil(t, err)
	m, _, _, err = invSrv.Create(ctxh, po)
	assert.Nil(t, err)
	newDM, err := m.PackCoreDocument()
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(accountID, newDM.Document.DocumentIdentifier))
	assert.True(t, testRepo().Exists(accountID, newDM.Document.CurrentVersion))
}

func TestService_DeriveInvoiceData(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()

	// some random model
	_, err := invSrv.DeriveInvoiceData(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := testingdocuments.CreateInvoicePayload()
	inv, err := invSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err, "must be non nil")
	data, err := invSrv.DeriveInvoiceData(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
}

func TestService_DeriveInvoiceResponse(t *testing.T) {
	// success
	invSrv := service{repo: testRepo()}
	payload := testingdocuments.CreateInvoicePayload()
	inv1, err := invSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.Nil(t, err, "must be non nil")
	inv, ok := inv1.(*Invoice)
	assert.True(t, ok)

	coreDoc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: []byte{},
	}
	coreDocModel := &documents.CoreDocumentModel{
		coreDoc,
		nil,
	}
	inv.CoreDocumentModel = coreDocModel
	resp, err := invSrv.DeriveInvoiceResponse(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, resp, "data must be non nil")
	assert.Equal(t, resp.Data, payload.Data, "data mismatch")
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	_, invSrv := getServiceWithMockedLayers()
	// nil payload
	doc, err := invSrv.DeriveFromUpdatePayload(nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// nil payload data
	doc, err = invSrv.DeriveFromUpdatePayload(nil, &clientinvoicepb.InvoiceUpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// messed up identifier
	contextHeader := testingconfig.CreateAccountContext(t, cfg)
	payload := &clientinvoicepb.InvoiceUpdatePayload{Identifier: "some identifier", Data: &clientinvoicepb.InvoiceData{}}
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
	idC, _ := contextutil.Self(contextHeader)
	old := new(Invoice)
	err = old.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), idC.ID.String())
	assert.Nil(t, err)
	old.CoreDocumentModel.Document.DocumentIdentifier = id
	old.CoreDocumentModel.Document.CurrentVersion = id
	old.CoreDocumentModel.Document.DocumentRoot = utils.RandomSlice(32)
	err = testRepo().Create(accountID, id, old)
	assert.Nil(t, err)
	payload.Data = &clientinvoicepb.InvoiceData{
		Sender:      "0x010101010101",
		Recipient:   "0x010203040506",
		Payee:       "0x010203020406",
		GrossAmount: 42,
		ExtraData:   "some data",
		Currency:    "EUR",
	}
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))
	assert.Nil(t, doc)

	// failed core document new version
	payload.Data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	payload.Collaborators = []string{"some wrong ID"}
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPrepareCoreDocument, err))
	assert.Nil(t, doc)

	// success
	wantCollab := utils.RandomSlice(6)
	payload.Collaborators = []string{hexutil.Encode(wantCollab)}
	doc, err = invSrv.DeriveFromUpdatePayload(contextHeader, payload)
	assert.Nil(t, err)
	assert.NotNil(t, doc)
	dm, err := doc.PackCoreDocument()
	assert.Nil(t, err)
	assert.Equal(t, wantCollab, dm.Document.Collaborators[2])
	assert.Len(t, dm.Document.Collaborators, 3)
	oldDM, err := old.PackCoreDocument()
	assert.Nil(t, err)
	assert.Equal(t, oldDM.Document.DocumentIdentifier, dm.Document.DocumentIdentifier)
	assert.Equal(t, payload.Identifier, hexutil.Encode(dm.Document.DocumentIdentifier))
	assert.Equal(t, oldDM.Document.CurrentVersion, dm.Document.PreviousVersion)
	assert.Equal(t, oldDM.Document.NextVersion, dm.Document.CurrentVersion)
	assert.NotNil(t, dm.Document.NextVersion)
	assert.Equal(t, payload.Data, doc.(*Invoice).getClientData())
}

func TestService_Update(t *testing.T) {
	_, srv := getServiceWithMockedLayers()
	invSrv := srv.(service)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// pack failed
	model := &mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("pack error")).Once()
	_, _, _, err := invSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack error")

	// missing last version
	model = &mockModel{}
	dm := documents.NewCoreDocModel()
	model.On("PackCoreDocument").Return(dm, nil).Once()
	_, _, _, err = invSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")

	payload := testingdocuments.CreateInvoicePayload()
	payload.Collaborators = []string{"0x010203040506"}
	inv, err := invSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	dm, err = inv.PackCoreDocument()
	assert.Nil(t, err)
	dm.Document.DocumentRoot = utils.RandomSlice(32)
	inv.(*Invoice).CoreDocumentModel = dm
	testRepo().Create(accountID, dm.Document.CurrentVersion, inv)

	// calculate data root fails
	model = &mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Once()
	_, _, _, err = invSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// anchor fails
	data, err := invSrv.DeriveInvoiceData(inv)
	assert.Nil(t, err)
	data.GrossAmount = 100
	data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	collab := hexutil.Encode(utils.RandomSlice(6))
	newInv, err := invSrv.DeriveFromUpdatePayload(ctxh, &clientinvoicepb.InvoiceUpdatePayload{
		Identifier:    hexutil.Encode(dm.Document.DocumentIdentifier),
		Collaborators: []string{collab},
		Data:          data,
	})
	assert.Nil(t, err)
	newData, err := invSrv.DeriveInvoiceData(newInv)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
	inv, _, _, err = invSrv.Update(ctxh, newInv)
	assert.Nil(t, err)
	assert.NotNil(t, inv)
	newDM, err := inv.PackCoreDocument()
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(accountID, newDM.Document.DocumentIdentifier))
	assert.True(t, testRepo().Exists(accountID, newDM.Document.CurrentVersion))
	assert.True(t, testRepo().Exists(accountID, newDM.Document.PreviousVersion))
	newData, err = invSrv.DeriveInvoiceData(inv)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)

}

func TestService_calculateDataRoot(t *testing.T) {
	_, srv := getServiceWithMockedLayers()
	invSrv := srv.(service)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// type mismatch
	inv, err := invSrv.validateAndPersist(ctxh, nil, &testingdocuments.MockModel{}, nil)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	inv, err = invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateInvoicePayload())
	assert.Nil(t, err)
	assert.Nil(t, inv.(*Invoice).CoreDocumentModel.Document.DataRoot)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, v)
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	inv, err = invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateInvoicePayload())
	assert.Nil(t, err)
	assert.Nil(t, inv.(*Invoice).CoreDocumentModel.Document.DataRoot)
	err = invSrv.repo.Create(accountID, inv.(*Invoice).CoreDocumentModel.Document.CurrentVersion, inv)
	assert.Nil(t, err)
	inv, err = invSrv.validateAndPersist(ctxh, nil, inv, CreateValidator())
	assert.Nil(t, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db repository could not create the given model, key already exists")

	// success
	inv, err = invSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateInvoicePayload())
	assert.Nil(t, err)
	assert.Nil(t, inv.(*Invoice).CoreDocumentModel.Document.DataRoot)
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
		testRepoGlobal.Register(&Invoice{})
	}
	return testRepoGlobal
}
