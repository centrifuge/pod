// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/stretchr/testify/assert"
)

func CreateCDWithEmbeddedInvoice(t *testing.T, invoiceData invoicepb.InvoiceData) *documents.CoreDocumentModel {
	identifier := []byte("1")
	invoiceModel := &Invoice{}
	invoiceModel.CoreDocumentModel = documents.NewCoreDocModel()
	invoiceModel.loadFromP2PProtobuf(&invoiceData)
	_, err := invoiceModel.getInvoiceSalts(&invoiceData)
	assert.NoError(t, err)
	cdm, err := invoiceModel.PackCoreDocument()
	assert.NoError(t, err)
	cdm.Document.DocumentIdentifier = identifier

	return cdm
}
