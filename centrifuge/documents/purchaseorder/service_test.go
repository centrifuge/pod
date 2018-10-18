package purchaseorder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var poSrv = service{}

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

func TestService_CreateProofs(t *testing.T) {
	p, err := poSrv.CreateProofs(nil, nil)
	assert.Nil(t, p)
	assert.Error(t, err)
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

func TestService_GetLastVersion(t *testing.T) {
	m, err := poSrv.GetLastVersion(nil)
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
