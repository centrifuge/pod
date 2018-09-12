// +build unit

package purchaseorder

import (
	"reflect"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
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

	purchaseorderDoc := Empty()
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

	generatedCoreDocument, err := purchaseorderDoc.ConvertToCoreDocument()
	assert.Nil(t, err, "Error converting to CoreDocument")
	generatedCoreDocumentBytes, err := proto.Marshal(generatedCoreDocument)
	assert.Nil(t, err, "Error marshaling generatedCoreDocument")

	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	assert.Nil(t, err, "Error marshaling coreDocument")
	assert.Equal(t, coreDocumentBytes, generatedCoreDocumentBytes,
		"Generated & converted documents are not identical")

	convertedPurchaseOrderDoc, err := NewFromCoreDocument(generatedCoreDocument)
	convertedGeneratedPurchaseOrderDoc, err := NewFromCoreDocument(generatedCoreDocument)
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
	po, err := NewFromCoreDocument(nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, po, "document should be nil")
}

func TestNewInvoice_NilDocument(t *testing.T) {
	po, err := New(nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, po, "document should be nil")
}

func TestValidate(t *testing.T) {
	type want struct {
		valid  bool
		errMsg string
		errs   map[string]string
	}

	var (
		id1 = tools.RandomSlice(32)
		id2 = tools.RandomSlice(32)
		id3 = tools.RandomSlice(32)
		id4 = tools.RandomSlice(32)
		id5 = tools.RandomSlice(32)
	)

	validCoreDoc := &coredocumentpb.CoreDocument{
		DocumentRoot:       id1,
		DocumentIdentifier: id2,
		CurrentIdentifier:  id3,
		NextIdentifier:     id4,
		DataRoot:           id5,
		CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
			DocumentIdentifier: id1,
			CurrentIdentifier:  id2,
			NextIdentifier:     id3,
			DataRoot:           id4,
			PreviousRoot:       id5,
		},
	}

	tests := []struct {
		po   *purchaseorderpb.PurchaseOrderDocument
		want want
	}{
		{
			po: nil,
			want: want{
				valid:  false,
				errMsg: centerrors.NilDocument,
			},
		},

		{
			po: &purchaseorderpb.PurchaseOrderDocument{},
			want: want{
				valid:  false,
				errMsg: centerrors.NilDocument,
			},
		},

		{
			po: &purchaseorderpb.PurchaseOrderDocument{CoreDocument: validCoreDoc},
			want: want{
				valid:  false,
				errMsg: centerrors.NilDocumentData,
			},
		},

		{
			po: &purchaseorderpb.PurchaseOrderDocument{
				CoreDocument: validCoreDoc,
				Data: &purchaseorderpb.PurchaseOrderData{
					PoNumber:         "po1234",
					OrderName:        "Jack",
					OrderZipcode:     "921007",
					OrderCountry:     "AUS",
					RecipientName:    "John",
					RecipientZipcode: "12345",
					Currency:         "EUR",
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid Purchase Order",
				errs: map[string]string{
					"po_recipient_country": centerrors.RequiredField,
					"po_order_amount":      centerrors.RequirePositiveNumber,
					"po_salts":             centerrors.RequiredField,
				},
			},
		},

		{
			po: &purchaseorderpb.PurchaseOrderDocument{
				CoreDocument: validCoreDoc,
				Data: &purchaseorderpb.PurchaseOrderData{
					PoNumber:         "po1234",
					OrderName:        "Jack",
					OrderZipcode:     "921007",
					OrderCountry:     "Australia",
					RecipientName:    "John",
					RecipientZipcode: "12345",
					RecipientCountry: "Germany",
					Currency:         "EUR",
					OrderAmount:      800,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid Purchase Order",
				errs: map[string]string{
					"po_salts": centerrors.RequiredField,
				},
			},
		},

		{
			po: &purchaseorderpb.PurchaseOrderDocument{
				CoreDocument: validCoreDoc,
				Data: &purchaseorderpb.PurchaseOrderData{
					PoNumber:         "po1234",
					OrderName:        "Jack",
					OrderZipcode:     "921007",
					OrderCountry:     "Australia",
					RecipientName:    "John",
					RecipientZipcode: "12345",
					RecipientCountry: "Germany",
					Currency:         "EUR",
					OrderAmount:      800,
				},
				Salts: &purchaseorderpb.PurchaseOrderDataSalts{},
			},
			want: want{valid: true},
		},
	}

	for _, c := range tests {
		got := want{}
		got.valid, got.errMsg, got.errs = Validate(c.po)
		if !reflect.DeepEqual(c.want, got) {
			t.Fatalf("%v != %v", c.want, got)
		}
	}
}
