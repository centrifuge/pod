// +build unit

package purchaseorder

import (
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestPurchaseOrderCoreDocumentConverter(t *testing.T) {
	identifier := []byte("1")
	purchaseorderData := purchaseorderpb.PurchaseOrderData{
		NetAmount: 100,
	}
	purchaseorderSalts := purchaseorderpb.PurchaseOrderDataSalts{}

	purchaseorderDoc := NewEmptyPurchaseOrder()
	purchaseorderDoc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
	}
	purchaseorderDoc.Document.Data = &purchaseorderData
	purchaseorderDoc.Document.Salts = &purchaseorderSalts

	serializedPurchaseOrder, err := proto.Marshal(&purchaseorderData)
	assert.Nil(t, err, "Could not serialize PurchaseOrderData")

	serializedSalts, err := proto.Marshal(&purchaseorderSalts)
	assert.Nil(t, err, "Could not serialize PurchaseOrderDataSalts")

	purchaseorderAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   serializedPurchaseOrder,
	}
	purchaseorderSaltsAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderSaltsTypeUrl,
		Value:   serializedSalts,
	}
	coreDocument := coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData:       &purchaseorderAny,
		EmbeddedDataSalts:  &purchaseorderSaltsAny,
	}

	generatedCoreDocument := purchaseorderDoc.ConvertToCoreDocument()
	generatedCoreDocumentBytes, err := proto.Marshal(generatedCoreDocument)
	assert.Nil(t, err, "Error marshaling generatedCoreDocument")

	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	assert.Nil(t, err, "Error marshaling coreDocument")
	assert.Equal(t, coreDocumentBytes, generatedCoreDocumentBytes,
		"Generated & converted documents are not identical")

	convertedPurchaseOrderDoc, err := NewPurchaseOrderFromCoreDocument(generatedCoreDocument)
	convertedGeneratedPurchaseOrderDoc, err := NewPurchaseOrderFromCoreDocument(generatedCoreDocument)
	purchaseorderBytes, err := proto.Marshal(purchaseorderDoc.Document)
	assert.Nil(t, err, "Error marshaling purchaseorderDoc")

	convertedGeneratedPurchaseOrderBytes, err := proto.Marshal(convertedGeneratedPurchaseOrderDoc.Document)
	assert.Nil(t, err, "Error marshaling convertedGeneratedPurchaseOrderDoc")

	convertedPurchaseOrderBytes, err := proto.Marshal(convertedPurchaseOrderDoc.Document)
	assert.Nil(t, err, "Error marshaling convertedGeneratedPurchaseOrderDoc")

	assert.Equal(t, purchaseorderBytes, convertedGeneratedPurchaseOrderBytes,
		"purchaseorderBytes and convertedGeneratedPurchaseOrderBytes do not match")
	assert.Equal(t, purchaseorderBytes, convertedPurchaseOrderBytes,
		"purchaseorderBytes and convertedPurchaseOrderBytes do not match")

}

func TestNewInvoiceFromCoreDocument_NilDocument(t *testing.T) {
	po, err := NewPurchaseOrderFromCoreDocument(nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, po, "document should be nil")
}

func TestNewInvoice_NilDocument(t *testing.T) {
	po, err := NewPurchaseOrder(nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, po, "document should be nil")
}
