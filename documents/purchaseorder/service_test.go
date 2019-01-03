// +build unit

package purchaseorder

import (
	"context"
	"math/big"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
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
	tenantID    = common.DummyIdentity.Bytes()
	centIDBytes = utils.RandomSlice(identity.CentIDLength)
	key1Pub     = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1        = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
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

func getServiceWithMockedLayers() (*testingcommons.MockIDService, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	idService := &testingcommons.MockIDService{}
	idService.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil)
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	txRepo := ctx[transactions.BootstrappedRepo].(transactions.Repository)
	return idService, DefaultService(c, testRepo(), &mockAnchorRepo{}, idService, queueSrv, txRepo)
}

func TestService_Update(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)

	// pack failed
	model := &testingdocuments.MockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("pack error")).Once()
	_, _, err = poSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack error")

	// missing last version
	model = &testingdocuments.MockModel{}
	cd := coredocument.New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	_, _, err = poSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")

	payload := testingdocuments.CreatePOPayload()
	payload.Collaborators = []string{"0x010203040506"}
	po, err := poSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	cd, err = po.PackCoreDocument()
	assert.Nil(t, err)
	cd.DocumentRoot = utils.RandomSlice(32)
	po.(*PurchaseOrder).CoreDocument = cd
	testRepo().Create(tenantID, cd.CurrentVersion, po)

	// calculate data root fails
	model = &testingdocuments.MockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	_, _, err = poSrv.Update(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	data, err := poSrv.DerivePurchaseOrderData(po)
	assert.Nil(t, err)
	data.OrderAmount = 100
	data.ExtraData = hexutil.Encode(utils.RandomSlice(32))
	collab := hexutil.Encode(utils.RandomSlice(6))
	newInv, err := poSrv.DeriveFromUpdatePayload(ctxh, &clientpurchaseorderpb.PurchaseOrderUpdatePayload{
		Identifier:    hexutil.Encode(cd.DocumentIdentifier),
		Collaborators: []string{collab},
		Data:          data,
	})
	assert.Nil(t, err)
	newData, err := poSrv.DerivePurchaseOrderData(newInv)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
	po, _, err = poSrv.Update(ctxh, newInv)
	assert.Nil(t, err)
	assert.NotNil(t, po)

	newCD, err := po.PackCoreDocument()
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(tenantID, newCD.DocumentIdentifier))
	assert.True(t, testRepo().Exists(tenantID, newCD.CurrentVersion))
	assert.True(t, testRepo().Exists(tenantID, newCD.PreviousVersion))

	newData, err = poSrv.DerivePurchaseOrderData(po)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	poSrv := service{config: c, repo: testRepo()}

	// nil payload
	doc, err := poSrv.DeriveFromUpdatePayload(nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// nil payload data
	doc, err = poSrv.DeriveFromUpdatePayload(nil, &clientpurchaseorderpb.PurchaseOrderUpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))
	assert.Nil(t, doc)

	// messed up identifier
	contextHeader, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)
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
	old.CoreDocument.DocumentIdentifier = id
	old.CoreDocument.CurrentVersion = id
	old.CoreDocument.DocumentRoot = utils.RandomSlice(32)
	err = testRepo().Create(tenantID, id, old)
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
	cd, err := doc.PackCoreDocument()
	assert.Nil(t, err)
	assert.Equal(t, wantCollab, cd.Collaborators[1])
	assert.Len(t, cd.Collaborators, 2)
	oldCD, err := old.PackCoreDocument()
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
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)

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
	_, err := poSrv.DeriveFromCoreDocument(nil)
	assert.Error(t, err, "must fail to derive")

	// successful
	data := testingdocuments.CreatePOData()
	cd := testingdocuments.CreateCDWithEmbeddedPO(t, data)
	m, err := poSrv.DeriveFromCoreDocument(cd)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	po, ok := m.(*PurchaseOrder)
	assert.True(t, ok, "must be true")
	assert.Equal(t, po.Recipient[:], data.Recipient)
	assert.Equal(t, po.OrderAmount, data.OrderAmount)
}

func TestService_Create(t *testing.T) {
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	_, poSrv := getServiceWithMockedLayers()

	// calculate data root fails
	m, _, err := poSrv.Create(ctxh, &testingdocuments.MockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// anchor fails
	po, err := poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	m, _, err = poSrv.Create(ctxh, po)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	newCD, err := m.PackCoreDocument()
	assert.Nil(t, err)
	assert.True(t, testRepo().Exists(tenantID, newCD.DocumentIdentifier))
	assert.True(t, testRepo().Exists(tenantID, newCD.CurrentVersion))
}

func createAnchoredMockDocument(t *testing.T, skipSave bool) (*PurchaseOrder, error) {
	i := &PurchaseOrder{
		PoNumber:     "test_po",
		OrderAmount:  42,
		CoreDocument: coredocument.New(),
	}
	err := i.calculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the purchase order
	corDoc, err := i.PackCoreDocument()
	if err != nil {
		return nil, err
	}
	assert.Nil(t, coredocument.FillSalts(corDoc))
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}

	centID, err := identity.ToCentID(centIDBytes)
	assert.Nil(t, err)
	signKey := identity.IDKey{
		PublicKey:  key1Pub[:],
		PrivateKey: key1,
	}
	idConfig := &identity.IDConfig{
		ID: centID,
		Keys: map[int]identity.IDKey{
			identity.KeyPurposeSigning: signKey,
		},
	}

	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, corDoc.SigningRoot)

	corDoc.Signatures = append(corDoc.Signatures, sig)

	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = i.UnpackCoreDocument(corDoc)
	if err != nil {
		return nil, err
	}

	if !skipSave {
		err = testRepo().Create(tenantID, i.CoreDocument.CurrentVersion, i)
		if err != nil {
			return nil, err
		}
	}

	return i, nil
}

// Functions returns service mocks
func mockSignatureCheck(i *PurchaseOrder, srv *testingcommons.MockIDService, poSrv Service) *testingcommons.MockIDService {
	idkey := &identity.EthereumIdentityKey{
		Key:       key1Pub,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	anchorID, _ := anchors.ToAnchorID(i.CoreDocument.DocumentIdentifier)
	docRoot, _ := anchors.ToDocumentRoot(i.CoreDocument.DocumentRoot)
	mockRepo := poSrv.(service).anchorRepository.(*mockAnchorRepo)
	mockRepo.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	id := &testingcommons.MockID{}
	centID, _ := identity.ToCentID(centIDBytes)
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	return srv
}

func TestService_CreateProofs(t *testing.T) {
	idService, poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, poSrv)
	proof, err := poSrv.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"po.po_number"})
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.DocumentID)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.VersionID)
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetReadableName(), "po.po_number")
}

func TestService_CreateProofsValidationFails(t *testing.T) {
	idService, poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	i.CoreDocument.SigningRoot = nil
	err = testRepo().Update(tenantID, i.CoreDocument.CurrentVersion, i)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, poSrv)
	_, err = poSrv.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"po.po_number"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestService_CreateProofsInvalidField(t *testing.T) {
	idService, poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, poSrv)
	_, err = poSrv.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invalid_field"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentProof, err))
}

func TestService_CreateProofsDocumentDoesntExist(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	_, err := poSrv.CreateProofs(utils.RandomSlice(32), []string{"po.po_number"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
}

func updatedAnchoredMockDocument(t *testing.T, model *PurchaseOrder) (*PurchaseOrder, error) {
	model.OrderAmount = 50
	err := model.calculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the purchase order
	corDoc, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}
	// hacky update to version
	corDoc.CurrentVersion = corDoc.NextVersion
	corDoc.NextVersion = utils.RandomSlice(32)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = model.UnpackCoreDocument(corDoc)
	if err != nil {
		return nil, err
	}
	err = testRepo().Create(tenantID, model.CoreDocument.CurrentVersion, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func TestService_CreateProofsForVersion(t *testing.T) {
	idService, poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, poSrv)
	olderVersion := i.CoreDocument.CurrentVersion
	i, err = updatedAnchoredMockDocument(t, i)
	assert.Nil(t, err)
	proof, err := poSrv.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, olderVersion, []string{"po.po_number"})
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.DocumentID)
	assert.Equal(t, olderVersion, proof.VersionID)
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetReadableName(), "po.po_number")
}

func TestService_CreateProofsForVersionDocumentDoesntExist(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	_, poSrv := getServiceWithMockedLayers()
	assert.Nil(t, err)
	_, err = poSrv.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, utils.RandomSlice(32), []string{"po.po_number"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))
}

func TestService_DerivePurchaseOrderData(t *testing.T) {
	var m documents.Model
	_, poSrv := getServiceWithMockedLayers()
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)

	// unknown type
	m = &testingdocuments.MockModel{}
	d, err := poSrv.DerivePurchaseOrderData(m)
	assert.Nil(t, d)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	payload := testingdocuments.CreatePOPayload()
	m, err = poSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	d, err = poSrv.DerivePurchaseOrderData(m)
	assert.Nil(t, err)
	assert.Equal(t, d.Currency, payload.Data.Currency)
}

func TestService_DerivePurchaseOrderResponse(t *testing.T) {
	poSrv := service{}
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, errors.New("pack core document failed")).Once()
	r, err := poSrv.DerivePurchaseOrderResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack core document failed")

	// cent id failed
	cd := coredocument.New()
	cd.Collaborators = [][]byte{{1, 2, 3, 4, 5, 6}, {5, 6, 7}}
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	r, err = poSrv.DerivePurchaseOrderResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length byte slice provided for centID")

	// derive data failed
	cd.Collaborators = [][]byte{{1, 2, 3, 4, 5, 6}}
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	r, err = poSrv.DerivePurchaseOrderResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	payload := testingdocuments.CreatePOPayload()
	po, err := poSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, err)
	r, err = poSrv.DerivePurchaseOrderResponse(po)
	assert.Nil(t, err)
	assert.Equal(t, payload.Data, r.Data)
	assert.Equal(t, []string{"0x010101010101", "0x010101010101"}, r.Header.Collaborators)
}

func createMockDocument() (*PurchaseOrder, error) {
	documentIdentifier := utils.RandomSlice(32)
	nextIdentifier := utils.RandomSlice(32)

	model := &PurchaseOrder{
		PoNumber:    "test_po",
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     documentIdentifier,
			NextVersion:        nextIdentifier,
		},
	}
	err := testRepo().Create(tenantID, documentIdentifier, model)
	return model, err
}

func TestService_GetCurrentVersion(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	poSrv := service{config: c, repo: testRepo()}
	thirdIdentifier := utils.RandomSlice(32)
	doc, err := createMockDocument()
	assert.Nil(t, err)

	mod1, err := poSrv.GetCurrentVersion(doc.CoreDocument.DocumentIdentifier)
	assert.Nil(t, err)

	poLoad1, _ := mod1.(*PurchaseOrder)
	assert.Equal(t, poLoad1.CoreDocument.CurrentVersion, doc.CoreDocument.DocumentIdentifier)

	po2 := &PurchaseOrder{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: doc.CoreDocument.DocumentIdentifier,
			CurrentVersion:     doc.CoreDocument.NextVersion,
			NextVersion:        thirdIdentifier,
		},
	}

	err = testRepo().Create(tenantID, doc.CoreDocument.NextVersion, po2)
	assert.Nil(t, err)

	mod2, err := poSrv.GetCurrentVersion(doc.CoreDocument.DocumentIdentifier)
	assert.Nil(t, err)

	poLoad2, _ := mod2.(*PurchaseOrder)
	assert.Equal(t, poLoad2.CoreDocument.CurrentVersion, doc.CoreDocument.NextVersion)
	assert.Equal(t, poLoad2.CoreDocument.NextVersion, thirdIdentifier)
}

func TestService_GetVersion_invalid_version(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	poSrv := service{config: c, repo: testRepo()}
	currentVersion := utils.RandomSlice(32)

	po := &PurchaseOrder{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: utils.RandomSlice(32),
			CurrentVersion:     currentVersion,
		},
	}
	err := testRepo().Create(tenantID, currentVersion, po)
	assert.Nil(t, err)

	mod, err := poSrv.GetVersion(utils.RandomSlice(32), currentVersion)
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))
	assert.Nil(t, mod)
}

func TestService_GetVersion(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	poSrv := service{config: c, repo: testRepo()}
	documentIdentifier := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	po := &PurchaseOrder{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     currentVersion,
		},
	}
	err := testRepo().Create(tenantID, currentVersion, po)
	assert.Nil(t, err)

	mod, err := poSrv.GetVersion(documentIdentifier, currentVersion)
	assert.Nil(t, err)
	loadpo, _ := mod.(*PurchaseOrder)
	assert.Equal(t, loadpo.CoreDocument.CurrentVersion, currentVersion)
	assert.Equal(t, loadpo.CoreDocument.DocumentIdentifier, documentIdentifier)

	mod, err = poSrv.GetVersion(documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_Exists(t *testing.T) {
	_, poSrv := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	po := &PurchaseOrder{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     documentIdentifier,
		},
	}
	err := testRepo().Create(tenantID, documentIdentifier, po)
	assert.Nil(t, err)

	exists := poSrv.Exists(documentIdentifier)
	assert.True(t, exists, "purchase order should exist")

	exists = poSrv.Exists(utils.RandomSlice(32))
	assert.False(t, exists, "purchase order should not exist")

}

func TestService_ReceiveAnchoredDocument(t *testing.T) {
	poSrv := service{}
	err := poSrv.ReceiveAnchoredDocument(nil, nil)
	assert.Error(t, err)
}

func TestService_RequestDocumentSignature(t *testing.T) {
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)
	poSrv := service{}
	s, err := poSrv.RequestDocumentSignature(ctxh, nil)
	assert.Nil(t, s)
	assert.Error(t, err)
}

func TestService_calculateDataRoot(t *testing.T) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	poSrv := service{config: c, repo: testRepo()}
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)

	// type mismatch
	po, err := poSrv.calculateDataRoot(nil, &testingdocuments.MockModel{}, nil)
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	po, err = poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrder).CoreDocument.DataRoot)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return errors.New("validations fail")
	})
	po, err = poSrv.calculateDataRoot(nil, po, v)
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	po, err = poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrder).CoreDocument.DataRoot)
	err = poSrv.repo.Create(tenantID, po.(*PurchaseOrder).CoreDocument.CurrentVersion, po)
	assert.Nil(t, err)
	po, err = poSrv.calculateDataRoot(nil, po, CreateValidator())
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists)

	// success
	po, err = poSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrder).CoreDocument.DataRoot)
	po, err = poSrv.calculateDataRoot(nil, po, CreateValidator())
	assert.Nil(t, err)
	assert.NotNil(t, po)
	assert.NotNil(t, po.(*PurchaseOrder).CoreDocument.DataRoot)
}

var testRepoGlobal documents.Repository

func testRepo() documents.Repository {
	if testRepoGlobal == nil {
		ldb, err := storage.NewLevelDBStorage(storage.GetRandomTestStoragePath())
		if err != nil {
			panic(err)
		}
		testRepoGlobal = documents.NewDBRepository(storage.NewLevelDBRepository(ldb))
		testRepoGlobal.Register(&PurchaseOrder{})
	}
	return testRepoGlobal
}
