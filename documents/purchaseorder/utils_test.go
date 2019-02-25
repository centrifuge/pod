// +build unit

package purchaseorder

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/stretchr/testify/assert"
)

func CreateCDWithEmbeddedPO(t *testing.T, poData purchaseorderpb.PurchaseOrderData) *documents.CoreDocumentModel {
	identifier := []byte("1")
	poModel := &PurchaseOrder{}
	poModel.CoreDocumentModel = documents.NewCoreDocModel()
	poModel.loadFromP2PProtobuf(&poData)
	_, err := poModel.getPurchaseOrderSalts(&poData)
	assert.NoError(t, err)
	cdm, err := poModel.PackCoreDocument()
	assert.NoError(t, err)
	cdm.Document.DocumentIdentifier = identifier

	return cdm
}
