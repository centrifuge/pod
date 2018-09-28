// +build unit

package invoice

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func createInvoiceData() invoicepb.InvoiceData {
	return invoicepb.InvoiceData{
		Recipient:   tools.RandomSlice(identity.CentIDByteLength),
		Sender:      tools.RandomSlice(identity.CentIDByteLength),
		Payee:       tools.RandomSlice(identity.CentIDByteLength),
		GrossAmount: 42,
	}
}

func createCDWithEmbeddedInvoice(t *testing.T, invoiceData invoicepb.InvoiceData) *coredocumentpb.CoreDocument {
	identifier := []byte("1")
	invoiceSalts := invoicepb.InvoiceDataSalts{}

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
	coreDocument := &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData:       &invoiceAny,
		EmbeddedDataSalts:  &invoiceSaltsAny,
	}
	return coreDocument
}

func TestInvoice_FromCoreDocuments_invalidParameter(t *testing.T) {

	invoiceModel := &InvoiceModel{}

	emptyCoreDocument := &coredocumentpb.CoreDocument{}
	err := invoiceModel.FromCoreDocument(emptyCoreDocument)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	err = invoiceModel.FromCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	err = invoiceModel.FromCoreDocument(coreDocument)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestInvoice_InitCoreDocument_successful(t *testing.T) {

	invoiceModel := &InvoiceModel{}

	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())
	err := invoiceModel.FromCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}

func TestInvoice_tInitCoreDocument_invalidCentId(t *testing.T) {

	invoiceModel := &InvoiceModel{}

	coreDocument := createCDWithEmbeddedInvoice(t, invoicepb.InvoiceData{
		Recipient:   tools.RandomSlice(identity.CentIDByteLength + 1),
		Sender:      tools.RandomSlice(identity.CentIDByteLength),
		Payee:       tools.RandomSlice(identity.CentIDByteLength),
		GrossAmount: 42,
	})
	err := invoiceModel.FromCoreDocument(coreDocument)
	assert.Error(t, err, "invalid centID should produce an error")

}

func TestInvoice_CoreDocument_successful(t *testing.T) {

	invoiceModel := &InvoiceModel{}

	//init model with a coreDocument
	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())
	invoiceModel.FromCoreDocument(coreDocument)

	returnedCoreDocument, err := invoiceModel.CoreDocument()
	assert.Nil(t, err, "transformation from invoice to coreDocument failed")

	assert.Equal(t, coreDocument.EmbeddedData, returnedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, returnedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")

}

func TestInvoice_ModelInterface(t *testing.T) {

	var i interface{} = &InvoiceModel{}
	_, ok := i.(documents.Model)
	assert.True(t, ok, "model interface not implemented correctly for invoiceModel")
}

func TestInvoice_Type(t *testing.T) {

	var model documents.Model
	model = &InvoiceModel{}

	assert.Equal(t, model.Type(), reflect.TypeOf(&InvoiceModel{}), "InvoiceType not correct")

}

func TestInvoice_JSON(t *testing.T) {
	invoiceModel := &InvoiceModel{}

	//init model with a coreDocument
	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())
	invoiceModel.FromCoreDocument(coreDocument)

	jsonBytes, err := invoiceModel.JSON()

	assert.Nil(t, err, "marshal to json didn't work correctly")

	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	err = invoiceModel.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	recievedCoreDocument, err := invoiceModel.CoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, recievedCoreDocument.EmbeddedData, coreDocument.EmbeddedData, "JSON unmarshal damaged invoice variables")

}
