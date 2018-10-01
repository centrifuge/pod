// +build unit

// Important
// Note: After the migration to the new invoice model this file will not exist anymore

package invoice

import (
	"flag"
	"os"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	db := cc.GetLevelDBStorage()
	InitLegacyRepository(db)
	coredocumentrepository.InitLevelDBRepository(db)
	invService = &service{repo: GetRepository()}
	flag.Parse()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

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

	generatedCoreDocument, err := invoiceDoc.ConvertToCoreDocument()
	assert.Nil(t, err, "error converting to coredocument")
	generatedCoreDocumentBytes, err := proto.Marshal(generatedCoreDocument)
	assert.Nil(t, err, "Error marshaling generatedCoreDocument")

	coreDocumentBytes, err := proto.Marshal(&coreDocument)
	assert.Nil(t, err, "Error marshaling CoreDoc")
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

	validCoreDoc := testingutils.GenerateCoreDocument()
	tests := []struct {
		inv  *invoicepb.InvoiceDocument
		want want
	}{
		{
			inv: nil,
			want: want{
				valid:  false,
				errMsg: centerrors.NilDocument,
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{},
			want: want{
				valid:  false,
				errMsg: centerrors.NilDocument,
			},
		},

		{
			inv: &invoicepb.InvoiceDocument{CoreDocument: validCoreDoc},
			want: want{
				valid:  false,
				errMsg: centerrors.NilDocumentData,
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
					"inv_salts": centerrors.RequiredField,
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
