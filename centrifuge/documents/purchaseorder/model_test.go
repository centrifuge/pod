package purchaseorder

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestPO_FromCoreDocuments_invalidParameter(t *testing.T) {
	poModel := &PurchaseOrderModel{}

	emptyCoreDocument := &coredocumentpb.CoreDocument{}
	err := poModel.UnpackCoreDocument(emptyCoreDocument)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	err = poModel.UnpackCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	err = poModel.UnpackCoreDocument(coreDocument)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestPO_InitCoreDocument_successful(t *testing.T) {
	poModel := &PurchaseOrderModel{}

	poData := testingdocuments.CreatePOData()

	coreDocument := testingdocuments.CreateCDWithEmbeddedPO(t, poData)
	err := poModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}

func TestPO_InitCoreDocument_invalidCentId(t *testing.T) {
	poModel := &PurchaseOrderModel{}

	coreDocument := testingdocuments.CreateCDWithEmbeddedPO(t, purchaseorderpb.PurchaseOrderData{
		Recipient: tools.RandomSlice(identity.CentIDLength + 1)})

	err := poModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err)
	assert.Nil(t, poModel.Recipient)
}

func TestPO_CoreDocument_successful(t *testing.T) {
	poModel := &PurchaseOrderModel{}

	//init model with a CoreDoc
	poData := testingdocuments.CreatePOData()

	coreDocument := testingdocuments.CreateCDWithEmbeddedPO(t, poData)
	poModel.UnpackCoreDocument(coreDocument)

	returnedCoreDocument, err := poModel.PackCoreDocument()
	assert.Nil(t, err, "transformation from purchase order to CoreDoc failed")

	assert.Equal(t, coreDocument.EmbeddedData, returnedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, returnedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}

func TestPO_ModelInterface(t *testing.T) {
	var i interface{} = &PurchaseOrderModel{}
	_, ok := i.(documents.Model)
	assert.True(t, ok, "model interface not implemented correctly for purchaseOrder model")
}

func TestPO_Type(t *testing.T) {
	var model documents.Model
	model = &PurchaseOrderModel{}
	assert.Equal(t, model.Type(), reflect.TypeOf(&PurchaseOrderModel{}), "purchaseOrder Type not correct")
}

func TestPurchaseOrder_JSON(t *testing.T) {
	poModel := &PurchaseOrderModel{}
	poData := testingdocuments.CreatePOData()
	coreDocument := testingdocuments.CreateCDWithEmbeddedPO(t, poData)
	poModel.UnpackCoreDocument(coreDocument)

	jsonBytes, err := poModel.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	err = poModel.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	receivedCoreDocument, err := poModel.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged purchase order variables")
	assert.Equal(t, receivedCoreDocument.EmbeddedData, coreDocument.EmbeddedData, "JSON unmarshal damaged invoice variables")
}

func TestInvoiceModel_UnpackCoreDocument(t *testing.T) {
	var model documents.Model = new(PurchaseOrderModel)
	var err error

	// nil core doc
	err = model.UnpackCoreDocument(nil)
	assert.Error(t, err, "unpack must fail")

	// embed data missing
	err = model.UnpackCoreDocument(new(coredocumentpb.CoreDocument))
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// successful
	coreDocument := testingdocuments.CreateCDWithEmbeddedPO(t, testingdocuments.CreatePOData())
	err = model.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid core document with embedded purchase order shouldn't produce an error")

	receivedCoreDocument, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded purchase order")

	assert.Equal(t, coreDocument.EmbeddedData, receivedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, receivedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}
