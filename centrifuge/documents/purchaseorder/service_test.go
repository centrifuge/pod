// +build unit

package purchaseorder

import (
	"context"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	poSrv Service
	centID     = utils.RandomSlice(identity.CentIDLength)
	key1Pub    = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1       = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
)

func TestService_Update(t *testing.T) {
	m, err := poSrv.Update(context.Background(), nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	m, err := poSrv.DeriveFromUpdatePayload(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	m, err := poSrv.DeriveFromCreatePayload(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	m, err := poSrv.DeriveFromCoreDocument(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_Create(t *testing.T) {
	m, err := poSrv.Create(context.Background(), nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func setIdentityService(idService identity.Service) {
	identity.IDService = idService
}

func createAnchoredMockDocument(t *testing.T, skipSave bool) (*PurchaseOrderModel, error) {
	i := &PurchaseOrderModel{
		PoNumber: "test_po",
		OrderAmount:   42,
		CoreDocument:  coredocument.New(),
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
	anchorRepository.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	srv := &testingcommons.MockIDService{}
	id := &testingcommons.MockID{}
	centID, _ := identity.ToCentID(centID)
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	return srv
}

func TestService_CreateProofs(t *testing.T) {
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

func TestService_CreateProofsForVersion(t *testing.T) {
	p, err := poSrv.CreateProofsForVersion(nil, nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestService_DerivePurchaseOrderData(t *testing.T) {
	d, err := poSrv.DerivePurchaseOrderData(nil)
	assert.Nil(t, d)
	assert.Error(t, err)
}

func TestService_DerivePurchaseOrderResponse(t *testing.T) {
	r, err := poSrv.DerivePurchaseOrderResponse(nil)
	assert.Nil(t, r)
	assert.Error(t, err)
}

func TestService_GetCurrentVersion(t *testing.T) {
	m, err := poSrv.GetCurrentVersion(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_GetVersion(t *testing.T) {
	m, err := poSrv.GetVersion(nil, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_ReceiveAnchoredDocument(t *testing.T) {
	err := poSrv.ReceiveAnchoredDocument(nil, nil)
	assert.Error(t, err)
}

func TestService_RequestDocumentSignature(t *testing.T) {
	s, err := poSrv.RequestDocumentSignature(nil)
	assert.Nil(t, s)
	assert.Error(t, err)
}
