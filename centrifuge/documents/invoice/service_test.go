// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

var invService Service

func createPayload() *clientinvoicepb.InvoiceCreatePayload {
	return &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			GrossAmount: 42,
		},
	}
}

func TestService_DeriveFromCoreDocument_successful(t *testing.T) {
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
	var model documents.Model
	var err error

	model, err = invService.DeriveFromCoreDocument(coreDocument)
	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")

	receivedCoreDocument, err := model.CoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.Equal(t, coreDocument.EmbeddedData, receivedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, receivedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")

}

func TestService_DeriveWithCoreDocument_invalid(t *testing.T) {
	var err error
	_, err = invService.DeriveFromCoreDocument(nil)
	assert.Error(t, err, "core document equals nil should produce an error")
}

func TestService_DeriveFromPayload_successful(t *testing.T) {
	payload := createPayload()
	var model documents.Model
	var err error

	model, err = invService.DeriveFromPayload(payload)
	assert.Nil(t, err, "valid invoiceData shouldn't produce an error")

	receivedCoreDocument, err := model.CoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.NotNil(t, receivedCoreDocument.EmbeddedData, "embeddedData should be field")

}

func TestService_DeriveFromPayload_invalid(t *testing.T) {
	var err error
	_, err = invService.DeriveFromPayload(nil)
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")
}

func TestService_Create(t *testing.T) {
	payload := createPayload()
	inv, err := invService.DeriveFromPayload(payload)
	assert.Nil(t, err, "must be non nil")

	// successful creation
	err = invService.Create(inv)
	assert.Nil(t, err, "create must pass")

	coredoc, err := inv.CoreDocument()
	assert.Nil(t, err, "must be converted to coredocument")

	loadInv := new(InvoiceModel)
	err = GetRepository().LoadByID(coredoc.CurrentIdentifier, loadInv)
	assert.Nil(t, err, "Load must pass")
	assert.NotNil(t, loadInv, "must be non nil")

	invType := inv.(*InvoiceModel)
	assert.Equal(t, loadInv.GrossAmount, invType.GrossAmount)
	assert.Equal(t, loadInv.CoreDoc, invType.CoreDoc)

	// failed creation
	err = invService.Create(inv)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "document already exists")
}
