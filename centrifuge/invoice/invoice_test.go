// +build unit

package invoice

import (
	"reflect"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestInvoiceCoreDocumentConverter(t *testing.T) {
	identifier := []byte("1")
	invoiceData := invoicepb.InvoiceData{
		GrossAmount: 100,
	}
	invoiceSalts := invoicepb.InvoiceDataSalts{}

	invoiceDoc := Empty()
	invoiceDoc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
	}
	invoiceDoc.Document.Data = &invoiceData
	invoiceDoc.Document.Salts = &invoiceSalts

	serializedInvoice, err := proto.Marshal(&invoiceData)
	assert.Nil(t, err, "Could not serialize InvoiceData")

	serializedSalts, err := proto.Marshal(&invoiceSalts)
	assert.Nil(t, err, "Could not serialize InvoiceDataSalts")

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}
	invoiceSaltsAny := any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		Value:   serializedSalts,
	}
	coreDocument := coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData:       &invoiceAny,
		EmbeddedDataSalts:  &invoiceSaltsAny,
	}

	generatedCoreDocument := invoiceDoc.ConvertToCoreDocument()
	generatedCoreDocumentBytes, err := proto.Marshal(generatedCoreDocument)
	assert.Nil(t, err, "Error marshaling generatedCoreDocument")

	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	assert.Nil(t, err, "Error marshaling coreDocument")
	assert.Equal(t, coreDocumentBytes, generatedCoreDocumentBytes,
		"Generated & converted documents are not identical")

	convertedInvoiceDoc, err := NewFromCoreDocument(generatedCoreDocument)
	convertedGeneratedInvoiceDoc, err := NewFromCoreDocument(generatedCoreDocument)
	invoiceBytes, err := proto.Marshal(invoiceDoc.Document)
	assert.Nil(t, err, "Error marshaling invoiceDoc")

	convertedGeneratedInvoiceBytes, err := proto.Marshal(convertedGeneratedInvoiceDoc.Document)
	assert.Nil(t, err, "Error marshaling convertedGeneratedInvoiceDoc")

	convertedInvoiceBytes, err := proto.Marshal(convertedInvoiceDoc.Document)
	assert.Nil(t, err, "Error marshaling convertedGeneratedInvoiceDoc")

	assert.Equal(t, invoiceBytes, convertedGeneratedInvoiceBytes,
		"invoiceBytes and convertedGeneratedInvoiceBytes do not match")
	assert.Equal(t, invoiceBytes, convertedInvoiceBytes,
		"invoiceBytes and convertedInvoiceBytes do not match")

}

func TestNewInvoiceFromCoreDocument_NilDocument(t *testing.T) {
	inv, err := NewFromCoreDocument(nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, inv, "document should be nil")
}

func TestNewInvoice_NilDocument(t *testing.T) {
	inv, err := New(nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, inv, "document should be nil")
}

func TestValidate(t *testing.T) {
	type want struct {
		valid  bool
		errMsg string
		errs   map[string]string
	}

	var (
		id1 = tools.RandomSlice32()
		id2 = tools.RandomSlice32()
		id3 = tools.RandomSlice32()
		id4 = tools.RandomSlice32()
		id5 = tools.RandomSlice32()
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
		inv  *invoicepb.InvoiceDocument
		want want
	}{
		{
			inv: nil,
			want: want{
				valid:  false,
				errMsg: errors.NilDocument,
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{},
			want: want{
				valid:  false,
				errMsg: errors.NilDocument,
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{CoreDocument: validCoreDoc},
			want: want{
				valid:  false,
				errMsg: errors.NilDocumentData,
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{
				CoreDocument: validCoreDoc,
				Data: &invoicepb.InvoiceData{
					InvoiceNumber:    "inv1234",
					SenderName:       "Jack",
					SenderZipcode:    "921007",
					SenderCountry:    "AUS",
					RecipientName:    "John",
					RecipientZipcode: "12345",
					Currency:         "EUR",
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid Invoice",
				errs: map[string]string{
					"inv_recipient_country": errors.RequiredField,
					"inv_gross_amount":      errors.RequirePositiveNumber,
					"inv_salts":             errors.RequiredField,
				},
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{
				CoreDocument: validCoreDoc,
				Data: &invoicepb.InvoiceData{
					InvoiceNumber:    "inv1234",
					SenderName:       "Jack",
					SenderZipcode:    "921007",
					SenderCountry:    "AUS",
					RecipientName:    "John",
					RecipientZipcode: "12345",
					RecipientCountry: "Germany",
					Currency:         "EUR",
					GrossAmount:      800,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid Invoice",
				errs: map[string]string{
					"inv_salts": errors.RequiredField,
				},
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{
				CoreDocument: validCoreDoc,
				Data: &invoicepb.InvoiceData{
					InvoiceNumber:    "inv1234",
					SenderName:       "Jack",
					SenderZipcode:    "921007",
					SenderCountry:    "AUS",
					RecipientName:    "John",
					RecipientZipcode: "12345",
					RecipientCountry: "Germany",
					Currency:         "EUR",
					GrossAmount:      800,
				},
				Salts: &invoicepb.InvoiceDataSalts{},
			},
			want: want{valid: true},
		},
	}

	for _, c := range tests {
		got := want{}
		got.valid, got.errMsg, got.errs = Validate(c.inv)
		if !reflect.DeepEqual(c.want, got) {
			t.Fatalf("%v != %v", c.want, got)
		}
	}
}
