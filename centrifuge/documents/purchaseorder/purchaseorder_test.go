// +build unit

package purchaseorder

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&coredocumentrepository.Bootstrapper{},
		&Bootstrapper{},
	}
	anchorRepository := &mockAnchorRepo{}
	context := map[string]interface{}{
		bootstrap.BootstrappedAnchorRepository: anchorRepository,
	}
	bootstrap.RunTestBootstrappers(ibootstappers, context)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

type mockAnchorRepo struct {
	mock.Mock
	anchors.AnchorRepository
}

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
	po, err := New(nil, nil)

	assert.Error(t, err, "should have thrown an error")
	assert.Nil(t, po, "document should be nil")
}

func TestValidate(t *testing.T) {
	var (
		id1 = utils.RandomSlice(32)
		id2 = utils.RandomSlice(32)
		id3 = utils.RandomSlice(32)
		id4 = utils.RandomSlice(32)
		id5 = utils.RandomSlice(32)
	)

	validCoreDoc := &coredocumentpb.CoreDocument{
		DocumentRoot:       id1,
		DocumentIdentifier: id2,
		CurrentVersion:     id3,
		NextVersion:        id4,
		DataRoot:           id5,
		CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
			DocumentIdentifier: id1,
			CurrentVersion:     id2,
			NextVersion:        id3,
			DataRoot:           id4,
			PreviousRoot:       id5,
		},
	}

	tests := []struct {
		po  *purchaseorderpb.PurchaseOrderDocument
		msg string
	}{
		{
			po:  nil,
			msg: "nil document",
		},

		{
			po:  &purchaseorderpb.PurchaseOrderDocument{},
			msg: "nil document",
		},

		{
			po:  &purchaseorderpb.PurchaseOrderDocument{CoreDocument: validCoreDoc},
			msg: "missing purchase order data",
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
			msg: "missing purchase order salts",
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
		},
	}

	for _, c := range tests {
		err := Validate(c.po)
		if c.msg == "" {
			assert.Nil(t, err)
			continue
		}

		assert.Contains(t, err.Error(), c.msg)
	}
}
