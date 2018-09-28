// +build unit

package invoice

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

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

	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
	err := invoiceModel.FromCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}

func TestInvoice_tInitCoreDocument_invalidCentId(t *testing.T) {

	invoiceModel := &InvoiceModel{}

	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, invoicepb.InvoiceData{
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
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
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
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
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
