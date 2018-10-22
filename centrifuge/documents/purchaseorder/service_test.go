// +build unit

package purchaseorder

import (
	"context"
	"math/big"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"

	"github.com/stretchr/testify/assert"
)

var (
	centID  = utils.RandomSlice(identity.CentIDLength)
	key1Pub = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
)

func getServiceWithMockedLayers() Service {
	return DefaultService(getRepository(), &testingutils.MockCoreDocumentProcessor{}, mockAnchorRepository)
}

func TestService_Update(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	m, err := poSrv.Update(context.Background(), nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	m, err := poSrv.DeriveFromUpdatePayload(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	m, err := poSrv.DeriveFromCreatePayload(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	m, err := poSrv.DeriveFromCoreDocument(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_Create(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	m, err := poSrv.Create(context.Background(), nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func setIdentityService(idService identity.Service) {
	identity.IDService = idService
}

func createAnchoredMockDocument(t *testing.T, skipSave bool) (*PurchaseOrderModel, error) {
	i := &PurchaseOrderModel{
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
	coredocument.FillSalts(corDoc)
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}

	sig := signatures.Sign(&config.IdentityConfig{
		ID:         centID,
		PublicKey:  key1Pub[:],
		PrivateKey: key1,
	}, corDoc.SigningRoot)

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
		err = getRepository().Create(i.CoreDocument.CurrentVersion, i)
		if err != nil {
			return nil, err
		}
	}

	return i, nil
}

// Functions returns service mocks
func mockSignatureCheck(i *PurchaseOrderModel) identity.Service {
	idkey := &identity.EthereumIdentityKey{
		Key:       key1Pub,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	anchorID, _ := anchors.NewAnchorID(i.CoreDocument.DocumentIdentifier)
	docRoot, _ := anchors.NewDocRoot(i.CoreDocument.DocumentRoot)
	mockAnchorRepository.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	srv := &testingcommons.MockIDService{}
	id := &testingcommons.MockID{}
	centID, _ := identity.ToCentID(centID)
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	return srv
}

func TestService_CreateProofs(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	p, err := poSrv.CreateProofs(nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
	defer setIdentityService(identity.IDService)
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService := mockSignatureCheck(i)
	setIdentityService(idService)
	proof, err := poSrv.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"po_number"})
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.DocumentId)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.VersionId)
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetProperty(), "po_number")
}

func TestService_CreateProofsValidationFails(t *testing.T) {
	defer setIdentityService(identity.IDService)
	poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	i.CoreDocument.SigningRoot = nil
	err = getRepository().Update(i.CoreDocument.CurrentVersion, i)
	assert.Nil(t, err)
	idService := mockSignatureCheck(i)
	setIdentityService(idService)
	_, err = poSrv.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"po_number"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestService_CreateProofsInvalidField(t *testing.T) {
	defer setIdentityService(identity.IDService)
	poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService := mockSignatureCheck(i)
	setIdentityService(idService)
	_, err = poSrv.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invalid_field"})
	assert.Error(t, err)
	assert.Equal(t, "createProofs error No such field: invalid_field in obj", err.Error())
}

func TestService_CreateProofsDocumentDoesntExist(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	_, err := poSrv.CreateProofs(utils.RandomSlice(32), []string{"po_number"})
	assert.Error(t, err)
	assert.Equal(t, "document not found: leveldb: not found", err.Error())
}

func updatedAnchoredMockDocument(t *testing.T, model *PurchaseOrderModel) (*PurchaseOrderModel, error) {
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
	err = getRepository().Create(model.CoreDocument.CurrentVersion, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func TestService_CreateProofsForVersion(t *testing.T) {
	defer setIdentityService(identity.IDService)
	poSrv := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService := mockSignatureCheck(i)
	setIdentityService(idService)
	olderVersion := i.CoreDocument.CurrentVersion
	i, err = updatedAnchoredMockDocument(t, i)
	assert.Nil(t, err)
	proof, err := poSrv.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, olderVersion, []string{"po_number"})
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.DocumentId)
	assert.Equal(t, olderVersion, proof.VersionId)
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetProperty(), "po_number")
}

func TestService_CreateProofsForVersionDocumentDoesntExist(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	poSrv := getServiceWithMockedLayers()
	assert.Nil(t, err)
	_, err = poSrv.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, utils.RandomSlice(32), []string{"po_number"})
	assert.Error(t, err)
	assert.Equal(t, "document not found for the given version: leveldb: not found", err.Error())
}

func TestService_DerivePurchaseOrderData(t *testing.T) {

	poSrv := getServiceWithMockedLayers()
	d, err := poSrv.DerivePurchaseOrderData(nil)
	assert.Nil(t, d)
	assert.Error(t, err)
}

func TestService_DerivePurchaseOrderResponse(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	r, err := poSrv.DerivePurchaseOrderResponse(nil)
	assert.Nil(t, r)
	assert.Error(t, err)
}

func createMockDocument() (*PurchaseOrderModel, error) {
	documentIdentifier := utils.RandomSlice(32)
	nextIdentifier := utils.RandomSlice(32)

	model := &PurchaseOrderModel{
		PoNumber:    "test_po",
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     documentIdentifier,
			NextVersion:        nextIdentifier,
		},
	}
	err := getRepository().Create(documentIdentifier, model)
	return model, err
}

func TestService_GetCurrentVersion(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	thirdIdentifier := utils.RandomSlice(32)
	doc, err := createMockDocument()
	assert.Nil(t, err)

	mod1, err := poSrv.GetCurrentVersion(doc.CoreDocument.DocumentIdentifier)
	assert.Nil(t, err)

	invLoad1, _ := mod1.(*PurchaseOrderModel)
	assert.Equal(t, invLoad1.CoreDocument.CurrentVersion, doc.CoreDocument.DocumentIdentifier)

	inv2 := &PurchaseOrderModel{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: doc.CoreDocument.DocumentIdentifier,
			CurrentVersion:     doc.CoreDocument.NextVersion,
			NextVersion:        thirdIdentifier,
		},
	}

	err = getRepository().Create(doc.CoreDocument.NextVersion, inv2)
	assert.Nil(t, err)

	mod2, err := poSrv.GetCurrentVersion(doc.CoreDocument.DocumentIdentifier)
	assert.Nil(t, err)

	invLoad2, _ := mod2.(*PurchaseOrderModel)
	assert.Equal(t, invLoad2.CoreDocument.CurrentVersion, doc.CoreDocument.NextVersion)
	assert.Equal(t, invLoad2.CoreDocument.NextVersion, thirdIdentifier)
}

func TestService_GetVersion_invalid_version(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	currentVersion := utils.RandomSlice(32)

	inv := &PurchaseOrderModel{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: utils.RandomSlice(32),
			CurrentVersion:     currentVersion,
		},
	}
	err := getRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := poSrv.GetVersion(utils.RandomSlice(32), currentVersion)
	assert.EqualError(t, err, "[4]document not found for the given version: version is not valid for this identifier")
	assert.Nil(t, mod)
}

func TestService_GetVersion(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	inv := &PurchaseOrderModel{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     currentVersion,
		},
	}
	err := getRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := poSrv.GetVersion(documentIdentifier, currentVersion)
	assert.Nil(t, err)
	loadInv, _ := mod.(*PurchaseOrderModel)
	assert.Equal(t, loadInv.CoreDocument.CurrentVersion, currentVersion)
	assert.Equal(t, loadInv.CoreDocument.DocumentIdentifier, documentIdentifier)

	mod, err = poSrv.GetVersion(documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_ReceiveAnchoredDocument(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	err := poSrv.ReceiveAnchoredDocument(nil, nil)
	assert.Error(t, err)
}

func TestService_RequestDocumentSignature(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	s, err := poSrv.RequestDocumentSignature(nil)
	assert.Nil(t, s)
	assert.Error(t, err)
}
