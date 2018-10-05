// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
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
			ExtraData:   "0x",
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

func TestService_GetLastVersion(t *testing.T) {
	documentIdentifier := tools.RandomSlice(32)
	nextIdentifier := tools.RandomSlice(32)
	thirdIdentifier := tools.RandomSlice(32)
	inv1 := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentIdentifier:  documentIdentifier,
			NextIdentifier:     nextIdentifier,
		},
	}
	err := GetRepository().Create(documentIdentifier, inv1)
	assert.Nil(t, err)

	mod1, err := invService.GetLastVersion(documentIdentifier)
	assert.Nil(t, err)

	invLoad1, _ := mod1.(*InvoiceModel)
	assert.Equal(t, invLoad1.CoreDocument.CurrentIdentifier, documentIdentifier)

	inv2 := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentIdentifier:  nextIdentifier,
			NextIdentifier:     thirdIdentifier,
		},
	}

	err = GetRepository().Create(nextIdentifier, inv2)
	assert.Nil(t, err)

	mod2, err := invService.GetLastVersion(documentIdentifier)
	assert.Nil(t, err)

	invLoad2, _ := mod2.(*InvoiceModel)
	assert.Equal(t, invLoad2.CoreDocument.CurrentIdentifier, nextIdentifier)
	assert.Equal(t, invLoad2.CoreDocument.NextIdentifier, thirdIdentifier)
}

func TestService_GetVersion(t *testing.T) {
	documentIdentifier := tools.RandomSlice(32)
	currentIdentifier := tools.RandomSlice(32)

	inv := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentIdentifier:  currentIdentifier,
		},
	}
	err := GetRepository().Create(currentIdentifier, inv)
	assert.Nil(t, err)

	mod, err := invService.GetVersion(documentIdentifier, currentIdentifier)
	assert.Nil(t, err)
	loadInv, _ := mod.(*InvoiceModel)
	assert.Equal(t, loadInv.CoreDocument.CurrentIdentifier, currentIdentifier)
	assert.Equal(t, loadInv.CoreDocument.DocumentIdentifier, documentIdentifier)

	mod, err = invService.GetVersion(documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_Create(t *testing.T) {
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

func TestService_DeriveInvoiceData(t *testing.T) {
	// some random model
	_, err := invService.DeriveInvoiceData(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := createPayload()
	inv, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")
	data, err := invService.DeriveInvoiceData(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
}

func TestService_DeriveInvoiceResponse(t *testing.T) {
	model := &mockModel{
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte{},
		},
	}
	// some random model
	_, err := invService.DeriveInvoiceResponse(model)
	assert.Error(t, err, "Derive must fail")

	// success
	payload := createPayload()
	inv1, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")
	inv, ok := inv1.(*InvoiceModel)
	assert.True(t, ok)
	inv.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: []byte{},
	}
	resp, err := invService.DeriveInvoiceResponse(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, resp, "data must be non nil")
	assert.Equal(t, resp.Data, payload.Data, "data mismatch")
}
