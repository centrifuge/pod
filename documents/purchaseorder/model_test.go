// +build unit

package purchaseorder

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity/ideth"

	"github.com/centrifuge/go-centrifuge/testingutils/testingtx"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/identity/ethid"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/queue"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var configService config.Service

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	txMan := &testingtx.MockTxManager{}
	ctx[transactions.BootstrappedService] = txMan
	done := make(chan bool)
	txMan.On("ExecuteWithinTX", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uuid.Nil, done, nil)

	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&queue.Bootstrapper{},
		&ethid.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", cid.String())
	configService = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestPO_FromCoreDocuments_invalidParameter(t *testing.T) {
	poModel := &PurchaseOrder{}

	emptyCoreDocModel := &documents.CoreDocumentModel{
		nil,
		nil,
	}
	err := poModel.UnpackCoreDocument(emptyCoreDocModel)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	err = poModel.UnpackCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	coreDocModel := &documents.CoreDocumentModel{
		coreDocument,
		nil,
	}
	err = poModel.UnpackCoreDocument(coreDocModel)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestPO_InitCoreDocument_successful(t *testing.T) {
	poModel := &PurchaseOrder{}

	poData := testingdocuments.CreatePOData()

	coreDocumentModel := CreateCDWithEmbeddedPO(t, poData)
	poModel.CoreDocumentModel = coreDocumentModel
	err := poModel.UnpackCoreDocument(coreDocumentModel)
	assert.Nil(t, err, "valid coredocumentmodel shouldn't produce an error")
}

func TestPO_InitCoreDocument_invalidCentId(t *testing.T) {
	poModel := &PurchaseOrder{}

	coreDocumentModel := CreateCDWithEmbeddedPO(t, purchaseorderpb.PurchaseOrderData{
		Recipient: utils.RandomSlice(identity.CentIDLength + 1)})
	poModel.CoreDocumentModel = coreDocumentModel
	err := poModel.UnpackCoreDocument(coreDocumentModel)
	assert.Nil(t, err)
	assert.Nil(t, poModel.Recipient)
}

func TestPO_CoreDocument_successful(t *testing.T) {
	poModel := &PurchaseOrder{}

	//init model with a CoreDoc
	poData := testingdocuments.CreatePOData()

	coreDocumentModel := CreateCDWithEmbeddedPO(t, poData)
	poModel.CoreDocumentModel = coreDocumentModel
	poModel.UnpackCoreDocument(coreDocumentModel)

	returnedCoreDocumentModel, err := poModel.PackCoreDocument()
	assert.Nil(t, err, "transformation from purchase order to CoreDoc failed")

	assert.Equal(t, coreDocumentModel.Document.EmbeddedData, returnedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocumentModel.Document.EmbeddedDataSalts, returnedCoreDocumentModel.Document.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}

func TestPO_ModelInterface(t *testing.T) {
	var i interface{} = &PurchaseOrder{}
	_, ok := i.(documents.Model)
	assert.True(t, ok, "model interface not implemented correctly for purchaseOrder model")
}

func TestPO_Type(t *testing.T) {
	var model documents.Model
	model = &PurchaseOrder{}
	assert.Equal(t, model.Type(), reflect.TypeOf(&PurchaseOrder{}), "purchaseOrder Type not correct")
}

func TestPO_JSON(t *testing.T) {
	poModel := &PurchaseOrder{}
	poData := testingdocuments.CreatePOData()
	coreDocumentModel := CreateCDWithEmbeddedPO(t, poData)
	poModel.CoreDocumentModel = coreDocumentModel
	poModel.UnpackCoreDocument(coreDocumentModel)

	jsonBytes, err := poModel.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	err = poModel.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	receivedCoreDocumentModel, err := poModel.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged purchase order variables")
	assert.Equal(t, receivedCoreDocumentModel.Document.EmbeddedData, coreDocumentModel.Document.EmbeddedData, "JSON unmarshal damaged purchase order variables")
}

func TestPOModel_UnpackCoreDocument(t *testing.T) {
	var model = new(PurchaseOrder)
	var err error

	// nil core doc
	err = model.UnpackCoreDocument(nil)
	assert.Error(t, err, "unpack must fail")

	// embed data missing
	err = model.UnpackCoreDocument(new(documents.CoreDocumentModel))
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// successful
	coreDocumentModel := CreateCDWithEmbeddedPO(t, testingdocuments.CreatePOData())
	model.CoreDocumentModel = coreDocumentModel
	err = model.UnpackCoreDocument(coreDocumentModel)
	assert.Nil(t, err, "valid core document with embedded purchase order shouldn't produce an error")

	receivedCoreDocumentModel, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded purchase order")

	assert.Equal(t, coreDocumentModel.Document.EmbeddedData, receivedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocumentModel.Document.EmbeddedDataSalts, receivedCoreDocumentModel.Document.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}

func TestPOModel_getClientData(t *testing.T) {
	poData := testingdocuments.CreatePOData()
	poModel := new(PurchaseOrder)
	poModel.loadFromP2PProtobuf(&poData)

	data := poModel.getClientData()
	assert.NotNil(t, data, "purchase order data should not be nil")
	assert.Equal(t, data.OrderAmount, data.OrderAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, hexutil.Encode(poModel.Recipient[:]), "recipient should match")
}

func TestPOOrderModel_InitPOInput(t *testing.T) {
	id, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	// fail recipient
	data := &clientpurchaseorderpb.PurchaseOrderData{
		Recipient: "some recipient",
		ExtraData: "some data",
	}
	poModel := new(PurchaseOrder)
	err := poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data}, id.ID.String())
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "failed to decode extra data")
	assert.Nil(t, poModel.Recipient)
	assert.Nil(t, poModel.ExtraData)

	data.ExtraData = "0x010203020301"
	data.Recipient = "0x010203040506"

	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data}, id.ID.String())
	assert.Nil(t, err)
	assert.NotNil(t, poModel.ExtraData)
	assert.NotNil(t, poModel.Recipient)

	data.ExtraData = "0x010203020301"
	collabs := []string{"0x010102040506", "some id"}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs}, id.ID.String())
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collab1, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	collab2, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3")
	assert.NoError(t, err)
	collabs = []string{collab1.String(), collab2.String()}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs}, id.ID.String())
	assert.Nil(t, err, "must be nil")

	assert.Equal(t, poModel.Recipient[:], []byte{1, 2, 3, 4, 5, 6})
	assert.Equal(t, poModel.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})

	assert.Equal(t, poModel.CoreDocumentModel.Document.Collaborators, [][]byte{id.ID[:], collab1[:], collab2[:]})
}

func TestPOModel_calculateDataRoot(t *testing.T) {
	id, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	poModel := new(PurchaseOrder)
	err := poModel.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), id.ID.String())
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, poModel.PurchaseOrderSalts, "salts must be nil")

	dr, err := poModel.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
	assert.NotNil(t, poModel.PurchaseOrderSalts, "salts must be created")
}
func TestPOModel_createProofs(t *testing.T) {
	poModel, err := createMockPurchaseOrder(t)
	assert.Nil(t, err)
	proof, err := poModel.CreateProofs([]string{"po.po_number", "collaborators[0]", "document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := poModel.CoreDocumentModel.GetDocumentRootTree()

	// Validate po_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate collaborators[0]
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	assert.Equal(t, poModel.CoreDocumentModel.Document.Collaborators[0], proof[1].Value)

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestPOModel_createProofsFieldDoesNotExist(t *testing.T) {
	poModel, err := createMockPurchaseOrder(t)
	assert.Nil(t, err)
	_, err = poModel.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestPOModel_getDocumentDataTree(t *testing.T) {
	poModel := PurchaseOrder{PoNumber: "3213121", NetAmount: 2, OrderAmount: 2}
	tree, err := poModel.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("po.po_number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "po.po_number", leaf.Property.ReadableName())
}

func createMockPurchaseOrder(t *testing.T) (*PurchaseOrder, error) {
	poModel := &PurchaseOrder{PoNumber: "3213121", NetAmount: 2, OrderAmount: 2, Currency: "USD", CoreDocumentModel: documents.NewCoreDocModel()}
	poModel.CoreDocumentModel.Document.Collaborators = [][]byte{{1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}}
	dataRoot, err := poModel.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the purchaseOrder
	corDocModel, err := poModel.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	err = corDocModel.CalculateSigningRoot(dataRoot)
	if err != nil {
		return nil, err
	}
	err = corDocModel.CalculateDocumentRoot()
	if err != nil {
		return nil, err
	}
	poModel.UnpackCoreDocument(corDocModel)
	return poModel, nil
}
