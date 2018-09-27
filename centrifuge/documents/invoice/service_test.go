// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/stretchr/testify/assert"
)

func TestService_DeriveWithCoreDocument_successful(t *testing.T) {

	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())

	service := &Service{}

	var model documents.Model
	var err error

	model, err = service.DeriveWithCoreDocument(coreDocument)
	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")

	receivedCoreDocument, err := model.CoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.Equal(t, coreDocument.EmbeddedData, receivedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, receivedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")

}

func TestService_DeriveWithCoreDocument_invalid(t *testing.T) {

	service := &Service{}
	var err error

	_, err = service.DeriveWithCoreDocument(nil)
	assert.Error(t, err, "core document equals nil should produce an error")

}

func TestService_DeriveWithInvoiceInput_successful(t *testing.T) {

	invoiceInput := createInvoiceInput()

	service := &Service{}

	var model documents.Model
	var err error

	model, err = service.DeriveWithInvoiceInput(&invoiceInput)
	assert.Nil(t, err, "valid invoiceData shouldn't produce an error")

	receivedCoreDocument, err := model.CoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.NotNil(t, receivedCoreDocument.EmbeddedData, "embeddedData should be field")

}

func TestService_DeriveWithInvoiceInput_invalid(t *testing.T) {

	service := &Service{}
	var err error

	_, err = service.DeriveWithInvoiceInput(nil)
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")

}
