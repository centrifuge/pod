// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

var invService Service

func createPayload() *clientinvoicepb.InvoiceCreatePayload {
	return &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			Sender:      "0x010101010101",
			Recipient:   "0x010203040506",
			Payee:       "0x010203020406",
			GrossAmount: 42,
			Currency:    "EUR",
		},
	}
}

func TestDefaultService(t *testing.T) {
	srv := DefaultService(GetRepository())
	assert.NotNil(t, srv, "must be non-nil")
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	// nil doc
	_, err := invService.DeriveFromCoreDocument(nil)
	assert.Error(t, err, "must fail to derive")

	// successful
	data := testinginvoice.CreateInvoiceData()
	coreDoc := testinginvoice.CreateCDWithEmbeddedInvoice(t, data)
	model, err := invService.DeriveFromCoreDocument(coreDoc)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, model, "model must be non-nil")
	inv, ok := model.(*InvoiceModel)
	assert.True(t, ok, "must be true")
	assert.Equal(t, inv.Payee[:], data.Payee)
	assert.Equal(t, inv.Sender[:], data.Sender)
	assert.Equal(t, inv.Recipient[:], data.Recipient)
	assert.Equal(t, inv.GrossAmount, data.GrossAmount)
}

func TestService_DeriveFromPayload(t *testing.T) {
	payload := createPayload()
	var model documents.Model
	var err error

	// fail due to nil payload
	_, err = invService.DeriveFromCreatePayload(nil)
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")

	model, err = invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "valid invoiceData shouldn't produce an error")

	receivedCoreDocument, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")
	assert.NotNil(t, receivedCoreDocument.EmbeddedData, "embeddedData should be field")
}

func TestService_Create(t *testing.T) {
	// fail Validations
	err := invService.Create(&InvoiceModel{})
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "Invoice invalid")

	payload := createPayload()
	inv, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")

	// successful creation
	err = invService.Create(inv)
	assert.Nil(t, err, "create must pass")

	coredoc, err := inv.PackCoreDocument()
	assert.Nil(t, err, "must be converted to coredocument")

	loadInv := new(InvoiceModel)
	err = GetRepository().LoadByID(coredoc.CurrentIdentifier, loadInv)
	assert.Nil(t, err, "Load must pass")
	assert.NotNil(t, loadInv, "must be non nil")

	invType := inv.(*InvoiceModel)
	assert.Equal(t, loadInv.GrossAmount, invType.GrossAmount)
	assert.Equal(t, loadInv.CoreDocument, invType.CoreDocument)

	// failed creation
	err = invService.Create(inv)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "document already exists")
}

func TestService_DeriveCreateResponse(t *testing.T) {
	// some random model
	_, err := invService.DeriveCreateResponse(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := createPayload()
	inv, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")
	data, err := invService.DeriveCreateResponse(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
	assert.Equal(t, data, payload.Data, "data mismatch")
}

func TestService_SaveState(t *testing.T) {
	// unknown type
	err := invService.SaveState(&mockModel{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document of invalid type")

	inv := new(InvoiceModel)
	err = inv.InitInvoiceInput(createPayload())
	assert.Nil(t, err)

	// save state must fail missing core document
	err = invService.SaveState(inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "core document missing")

	// update fail
	coreDoc, err := inv.PackCoreDocument()
	assert.Nil(t, err)
	assert.NotNil(t, coreDoc)
	err = invService.SaveState(inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document doesn't exists")

	// successful
	err = GetRepository().Create(coreDoc.CurrentIdentifier, inv)
	assert.Nil(t, err)
	assert.Equal(t, inv.Currency, "EUR")
	assert.Nil(t, inv.CoreDocument.DataRoot)

	inv.Currency = "INR"
	inv.CoreDocument.DataRoot = tools.RandomSlice(32)
	err = invService.SaveState(inv)
	assert.Nil(t, err)

	loadInv := new(InvoiceModel)
	err = GetRepository().LoadByID(coreDoc.CurrentIdentifier, loadInv)
	assert.Nil(t, err)
	assert.Equal(t, loadInv.Currency, inv.Currency)
	assert.Equal(t, loadInv.CoreDocument.DataRoot, inv.CoreDocument.DataRoot)
}
