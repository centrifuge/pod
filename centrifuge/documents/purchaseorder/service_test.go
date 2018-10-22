package purchaseorder

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestService_Update(t *testing.T) {
	poSrv := service{}
	m, err := poSrv.Update(context.Background(), nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	poSrv := service{}
	m, err := poSrv.DeriveFromUpdatePayload(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	poSrv := service{}
	// nil payload
	m, err := poSrv.DeriveFromCreatePayload(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input is nil")

	// Init fails
	payload := &clientpurchaseorderpb.PurchaseOrderCreatePayload{
		Data: &clientpurchaseorderpb.PurchaseOrderData{
			ExtraData: "some data",
		},
	}

	m, err = poSrv.DeriveFromCreatePayload(payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purchase order init failed")

	// success
	payload.Data.ExtraData = "0x01020304050607"
	m, err = poSrv.DeriveFromCreatePayload(payload)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	po := m.(*PurchaseOrderModel)
	assert.Equal(t, hexutil.Encode(po.ExtraData), payload.Data.ExtraData)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	poSrv := service{}
	m, err := poSrv.DeriveFromCoreDocument(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_Create(t *testing.T) {
	poSrv := service{}
	m, err := poSrv.Create(context.Background(), nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_CreateProofs(t *testing.T) {
	poSrv := service{}
	p, err := poSrv.CreateProofs(nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestService_CreateProofsForVersion(t *testing.T) {
	poSrv := service{}
	p, err := poSrv.CreateProofsForVersion(nil, nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestService_DerivePurchaseOrderData(t *testing.T) {
	var m documents.Model
	poSrv := service{}

	// unknown type
	m = &testingdocuments.MockModel{}
	d, err := poSrv.DerivePurchaseOrderData(m)
	assert.Nil(t, d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document of invalid type")

	// success
	payload := testingdocuments.CreatePOPayload()
	m, err = poSrv.DeriveFromCreatePayload(payload)
	assert.Nil(t, err)
	d, err = poSrv.DerivePurchaseOrderData(m)
	assert.Nil(t, err)
	assert.Equal(t, d.Currency, payload.Data.Currency)
}

func TestService_DerivePurchaseOrderResponse(t *testing.T) {
	poSrv := service{}

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, fmt.Errorf("pack core document failed")).Once()
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
	assert.Contains(t, err.Error(), "document of invalid type")

	// success
	payload := testingdocuments.CreatePOPayload()
	po, err := poSrv.DeriveFromCreatePayload(payload)
	assert.Nil(t, err)
	r, err = poSrv.DerivePurchaseOrderResponse(po)
	assert.Nil(t, err)
	assert.Equal(t, r.Data, payload.Data)
	assert.Equal(t, r.Header.Collaborators, payload.Collaborators)
}

func TestService_GetCurrentVersion(t *testing.T) {
	poSrv := service{}
	m, err := poSrv.GetCurrentVersion(nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_GetVersion(t *testing.T) {
	poSrv := service{}
	m, err := poSrv.GetVersion(nil, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestService_ReceiveAnchoredDocument(t *testing.T) {
	poSrv := service{}
	err := poSrv.ReceiveAnchoredDocument(nil, nil)
	assert.Error(t, err)
}

func TestService_RequestDocumentSignature(t *testing.T) {
	poSrv := service{}
	s, err := poSrv.RequestDocumentSignature(nil)
	assert.Nil(t, s)
	assert.Error(t, err)
}

func TestService_calculateDataRoot(t *testing.T) {
	poSrv := service{repo: getRepository()}

	// type mismatch
	po, err := poSrv.calculateDataRoot(nil, &testingdocuments.MockModel{}, nil)
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// failed validator
	po, err = poSrv.DeriveFromCreatePayload(testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrderModel).CoreDocument.DataRoot)
	v := documents.ValidatorFunc(func(_, _ documents.Model) error {
		return fmt.Errorf("validations fail")
	})
	po, err = poSrv.calculateDataRoot(nil, po, v)
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validations fail")

	// create failed
	po, err = poSrv.DeriveFromCreatePayload(testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrderModel).CoreDocument.DataRoot)
	err = poSrv.repo.Create(po.(*PurchaseOrderModel).CoreDocument.CurrentVersion, po)
	assert.Nil(t, err)
	po, err = poSrv.calculateDataRoot(nil, po, CreateValidator())
	assert.Nil(t, po)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document already exists")

	// success
	po, err = poSrv.DeriveFromCreatePayload(testingdocuments.CreatePOPayload())
	assert.Nil(t, err)
	assert.Nil(t, po.(*PurchaseOrderModel).CoreDocument.DataRoot)
	po, err = poSrv.calculateDataRoot(nil, po, CreateValidator())
	assert.Nil(t, err)
	assert.NotNil(t, po)
	assert.NotNil(t, po.(*PurchaseOrderModel).CoreDocument.DataRoot)

}
