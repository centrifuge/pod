

package invoice

import (
	"encoding/json"
	"github.com/centrifuge/go-centrifuge/centrifuge/models"
	"os"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.InitTestConfig()
	result := m.Run()
	os.Exit(result)
}


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

func TestInvoice_InitWithCoreDocuments_invalidParameter(t *testing.T) {

	invoiceModel := &Invoice{}

	emptyCoreDocument := &coredocumentpb.CoreDocument{}
	err := invoiceModel.InitWithCoreDocument(emptyCoreDocument)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	err = invoiceModel.InitWithCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	err = invoiceModel.InitWithCoreDocument(coreDocument)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestInvoice_InitCoreDocument_successful(t *testing.T) {

	invoiceModel := &Invoice{}

	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())
	err := invoiceModel.InitWithCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}

func TestInvoice_tInitCoreDocument_invalidCentId(t *testing.T) {

	invoiceModel := &Invoice{}

	coreDocument := createCDWithEmbeddedInvoice(t, invoicepb.InvoiceData{
		Recipient:   tools.RandomSlice(identity.CentIDByteLength + 1),
		Sender:      tools.RandomSlice(identity.CentIDByteLength),
		Payee:       tools.RandomSlice(identity.CentIDByteLength),
		GrossAmount: 42,
	})
	err := invoiceModel.InitWithCoreDocument(coreDocument)
	assert.Error(t, err, "invalid centID should produce an error")

}

func TestInvoice_CoreDocument_successful(t *testing.T) {

	invoiceModel := &Invoice{}

	//init model with a coreDocument
	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())
	invoiceModel.InitWithCoreDocument(coreDocument)

	returnedCoreDocument, err := invoiceModel.CoreDocument()
	assert.Nil(t, err, "transformation from invoice to coreDocument failed")

	assert.Equal(t, coreDocument.EmbeddedData, returnedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, returnedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")

}

func TestInvoice_ModelInterface(t *testing.T) {

	var i interface{} = &Invoice{}
	_, ok := i.(models.Model)
	assert.True(t, ok, "model interface not implemented correctly for invoiceModel")
}

func TestInvoice_Type(t *testing.T) {

	var model models.Model
	model = &Invoice{}

	assert.Equal(t, model.Type(), reflect.TypeOf(&Invoice{}), "InvoiceType not correct")

}

func TestInvoice_JSON(t *testing.T) {
	invoiceModel := &Invoice{}

	//init model with a coreDocument
	coreDocument := createCDWithEmbeddedInvoice(t, createInvoiceData())
	invoiceModel.InitWithCoreDocument(coreDocument)

	jsonBytes, err := invoiceModel.JSON()

	assert.Nil(t, err, "marshal to json didn't work correctly")

	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	err = invoiceModel.InitWithJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	recievedCoreDocument, err := invoiceModel.CoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, recievedCoreDocument.EmbeddedData, coreDocument.EmbeddedData, "JSON unmarshal damaged invoice variables")

}
