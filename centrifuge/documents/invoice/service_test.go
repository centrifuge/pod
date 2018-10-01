// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/stretchr/testify/assert"
)

func createInvoiceInput() InvoiceInput {
	return InvoiceInput{GrossAmount: 42}
}

func TestService_DeriveWithCoreDocument_successful(t *testing.T) {
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
	invoiceService := &service{}
	var model documents.Model
	var err error

	var modelDeriver ModelDeriver
	modelDeriver = invoiceService

	model, err = modelDeriver.DeriveWithCoreDocument(coreDocument)
	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")

	receivedCoreDocument, err := model.CoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.Equal(t, coreDocument.EmbeddedData, receivedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, receivedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")

}

func TestService_DeriveWithCoreDocument_invalid(t *testing.T) {
	invoiceService := &service{}
	var err error

	_, err = invoiceService.DeriveWithCoreDocument(nil)
	assert.Error(t, err, "core document equals nil should produce an error")
}

func TestService_DeriveWithInvoiceInput_successful(t *testing.T) {
	invoiceInput := createInvoiceInput()
	var modelDeriver ModelDeriver
	modelDeriver = &service{}
	var model documents.Model
	var err error

	model, err = modelDeriver.DeriveWithInvoiceInput(&invoiceInput)
	assert.Nil(t, err, "valid invoiceData shouldn't produce an error")

	receivedCoreDocument, err := model.CoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.NotNil(t, receivedCoreDocument.EmbeddedData, "embeddedData should be field")

}

func TestService_DeriveWithInvoiceInput_invalid(t *testing.T) {
	invoiceService := &service{}
	var err error

	_, err = invoiceService.DeriveWithInvoiceInput(nil)
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")
}
