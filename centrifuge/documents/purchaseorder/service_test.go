// +build unit

package purchaseorder

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"

	"github.com/stretchr/testify/assert"
)

func getServiceWithMockedLayers() Service {
	return DefaultService(getRepository(), &testingutils.MockCoreDocumentProcessor{})
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

func TestService_CreateProofs(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	p, err := poSrv.CreateProofs(nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestService_CreateProofsForVersion(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	p, err := poSrv.CreateProofsForVersion(nil, nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
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
	documentIdentifier := tools.RandomSlice(32)
	nextIdentifier := tools.RandomSlice(32)

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
	thirdIdentifier := tools.RandomSlice(32)
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
	currentVersion := tools.RandomSlice(32)

	inv := &PurchaseOrderModel{
		OrderAmount: 42,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: tools.RandomSlice(32),
			CurrentVersion:     currentVersion,
		},
	}
	err := getRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := poSrv.GetVersion(tools.RandomSlice(32), currentVersion)
	assert.EqualError(t, err, "[4]document not found for the given version: version is not valid for this identifier")
	assert.Nil(t, mod)
}

func TestService_GetVersion(t *testing.T) {
	poSrv := getServiceWithMockedLayers()
	documentIdentifier := tools.RandomSlice(32)
	currentVersion := tools.RandomSlice(32)

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
