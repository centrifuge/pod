// +build unit

package purchaseorder

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ethid"
	"github.com/centrifuge/go-centrifuge/p2p"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/testingtx"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	txMan.On("ExecuteWithinTX", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(transactions.NilTxID(), done, nil)

	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&queue.Bootstrapper{},
		&ethid.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", cid.String())
	configService = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestPurchaseOrder_PackCoreDocument(t *testing.T) {
	id, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)

	po := new(PurchaseOrder)
	assert.Error(t, po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), id.ID.String()))

	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
	assert.NotNil(t, cd.EmbeddedDataSalts)
}

func TestPurchaseOrder_JSON(t *testing.T) {
	po := new(PurchaseOrder)
	id, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)
	assert.NoError(t, po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), id.ID.String()))

	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := po.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	po = new(PurchaseOrder)
	err = po.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := po.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, cd, ncd)
}

//func TestPOModel_UnpackCoreDocument(t *testing.T) {
//	var model = new(PurchaseOrder)
//	var err error
//
//	// nil core doc
//	err = model.UnpackCoreDocument(nil)
//	assert.Error(t, err, "unpack must fail")
//
//	// embed data missing
//	err = model.UnpackCoreDocument(new(documents.CoreDocumentModel))
//	assert.Error(t, err, "unpack must fail due to missing embed data")
//
//	// successful
//	coreDocumentModel := CreateCDWithEmbeddedPO(t, testingdocuments.CreatePOData())
//	model.CoreDocumentModel = coreDocumentModel
//	err = model.UnpackCoreDocument(coreDocumentModel)
//	assert.Nil(t, err, "valid core document with embedded purchase order shouldn't produce an error")
//
//	receivedCoreDocumentModel, err := model.PackCoreDocument()
//	assert.Nil(t, err, "model should be able to return the core document with embedded purchase order")
//
//	assert.Equal(t, coreDocumentModel.Document.EmbeddedData, receivedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be the same")
//	assert.Equal(t, coreDocumentModel.Document.EmbeddedDataSalts, receivedCoreDocumentModel.Document.EmbeddedDataSalts, "embeddedDataSalt should be the same")
//}

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

	collabs = []string{"0x010102040506", "0x010203020302"}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs}, id.ID.String())
	assert.Nil(t, err, "must be nil")

	assert.Equal(t, poModel.Recipient[:], []byte{1, 2, 3, 4, 5, 6})
	assert.Equal(t, poModel.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
}

func TestPOModel_calculateDataRoot(t *testing.T) {
	id, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	poModel := new(PurchaseOrder)
	err := poModel.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), id.ID.String())
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, poModel.PurchaseOrderSalts, "salts must be nil")

	dr, err := poModel.DataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
	assert.NotNil(t, poModel.PurchaseOrderSalts, "salts must be created")
}
func TestPOModel_GenerateProofs(t *testing.T) {
	po, err := createPurchaseOrder(t)
	assert.Nil(t, err)
	proof, err := po.CreateProofs([]string{"po.po_number", "collaborators[0]", "document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := po.DocumentRootTree()

	// Validate po_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate collaborators[0]
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	id, err := identity.ToCentID(proof[1].Value)
	assert.NoError(t, err)
	assert.True(t, po.CoreDocument.AccountCanRead(id))

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestPOModel_createProofsFieldDoesNotExist(t *testing.T) {
	poModel, err := createPurchaseOrder(t)
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

func createPurchaseOrder(t *testing.T) (*PurchaseOrder, error) {
	po := new(PurchaseOrder)
	po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), "0x010203040506")
	_, err := po.DataRoot()
	assert.NoError(t, err)
	_, err = po.SigningRoot()
	assert.NoError(t, err)
	_, err = po.DocumentRoot()
	assert.NoError(t, err)
	return po, nil
}
