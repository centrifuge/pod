// +build unit

package purchaseorder

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
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

func TestPO_JSON(t *testing.T) {
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
	assert.Equal(t, receivedCoreDocument.EmbeddedData, coreDocument.EmbeddedData, "JSON unmarshal damaged purchase order variables")
}

func TestPOModel_UnpackCoreDocument(t *testing.T) {
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

func TestPOModel_getClientData(t *testing.T) {
	poData := testingdocuments.CreatePOData()
	poModel := new(PurchaseOrderModel)
	poModel.loadFromP2PProtobuf(&poData)

	data := poModel.getClientData()
	assert.NotNil(t, data, "purchase order data should not be nil")
	assert.Equal(t, data.OrderAmount, data.OrderAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, hexutil.Encode(poModel.Recipient[:]), "recipient should match")
}

func TestPOOrderModel_InitPOInput(t *testing.T) {
	idConfig, err := ed25519keys.GetIDConfig()
	assert.Nil(t, err)
	self := idConfig.ID
	// fail recipient
	data := &clientpurchaseorderpb.PurchaseOrderData{
		Recipient: "some recipient",
		ExtraData: "some data",
	}
	poModel := new(PurchaseOrderModel)
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data})
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "failed to decode extra data")
	assert.Nil(t, poModel.Recipient)
	assert.Nil(t, poModel.ExtraData)

	data.ExtraData = "0x010203020301"
	data.Recipient = "0x010203040506"

	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data})
	assert.Nil(t, err)
	assert.NotNil(t, poModel.ExtraData)
	assert.NotNil(t, poModel.Recipient)

	data.ExtraData = "0x010203020301"
	collabs := []string{"0x010102040506", "some id"}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs})
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collabs = []string{"0x010102040506", "0x010203020302"}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs})
	assert.Nil(t, err, "must be nil")

	assert.Equal(t, poModel.Recipient[:], []byte{1, 2, 3, 4, 5, 6})
	assert.Equal(t, poModel.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
	assert.Equal(t, poModel.CoreDocument.Collaborators, [][]byte{self, {1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}})
}

func TestPOModel_calculateDataRoot(t *testing.T) {
	poModel := new(PurchaseOrderModel)
	err := poModel.InitPurchaseOrderInput(testingdocuments.CreatePOPayload())
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, poModel.PurchaseOrderSalt, "salts must be nil")

	err = poModel.calculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.NotNil(t, poModel.CoreDocument, "coredoc must be created")
	assert.NotNil(t, poModel.PurchaseOrderSalt, "salts must be created")
	assert.NotNil(t, poModel.CoreDocument.DataRoot, "data root must be filled")
}
func TestPOModel_createProofs(t *testing.T) {
	poModel, corDoc, err := createMockPurchaseOrder()
	assert.Nil(t, err)
	corDoc, proof, err := poModel.createProofs([]string{"po_number", "collaborators[0]", "document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	assert.NotNil(t, corDoc)
	tree, _ := coredocument.GetDocumentRootTree(corDoc)

	// Validate po_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate collaborators[0]
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestPOModel_createProofsFieldDoesNotExist(t *testing.T) {
	poModel, _, err := createMockPurchaseOrder()
	assert.Nil(t, err)
	_, _, err = poModel.createProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestPOModel_getDocumentDataTree(t *testing.T) {
	poModel := PurchaseOrderModel{PoNumber: "3213121", NetAmount: 2, OrderAmount: 2}
	tree, err := poModel.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("po_number")
	assert.Equal(t, "po_number", leaf.Property)
}

func createMockPurchaseOrder() (*PurchaseOrderModel, *coredocumentpb.CoreDocument, error) {
	poModel := &PurchaseOrderModel{PoNumber: "3213121", NetAmount: 2, OrderAmount: 2, Currency: "USD", CoreDocument: coredocument.New()}
	poModel.CoreDocument.Collaborators = [][]byte{{1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}}
	err := poModel.calculateDataRoot()
	if err != nil {
		return nil, nil, err
	}
	// get the coreDoc for the purchaseOrder
	corDoc, err := poModel.PackCoreDocument()
	if err != nil {
		return nil, nil, err
	}
	coredocument.FillSalts(corDoc)
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, nil, err
	}
	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, nil, err
	}
	poModel.UnpackCoreDocument(corDoc)
	return poModel, corDoc, nil
}
