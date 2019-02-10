package purchaseorder

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/stretchr/testify/assert"
)

func CreateCDWithEmbeddedPO(t *testing.T, poData purchaseorderpb.PurchaseOrderData) *coredocumentpb.CoreDocument {
	identifier := []byte("1")
	poModel := &PurchaseOrder{}
	poModel.loadFromP2PProtobuf(&poData)
	_, err := poModel.getPurchaseOrderSalts(&poData)
	assert.NoError(t, err)
	coreDocument, err := poModel.PackCoreDocument()
	assert.NoError(t, err)
	coreDocument.DocumentIdentifier = identifier

	return coreDocument
}
