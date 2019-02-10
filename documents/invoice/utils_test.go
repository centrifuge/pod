package invoice

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/stretchr/testify/assert"
)

func CreateCDWithEmbeddedInvoice(t *testing.T, invoiceData invoicepb.InvoiceData) *coredocumentpb.CoreDocument {
	identifier := []byte("1")
	invoiceModel := &Invoice{}
	invoiceModel.loadFromP2PProtobuf(&invoiceData)
	_, err := invoiceModel.getInvoiceSalts(&invoiceData)
	assert.NoError(t, err)
	coreDocument, err := invoiceModel.PackCoreDocument()
	assert.NoError(t, err)
	coreDocument.DocumentIdentifier = identifier

	return coreDocument
}
