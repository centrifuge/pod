// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
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
