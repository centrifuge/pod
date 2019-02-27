// +build unit

package purchaseorder

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/nft"
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
	"github.com/golang/protobuf/ptypes/any"
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
	ctx[nft.BootstrappedPayObService] = new(testingdocuments.MockRegistry)
	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&queue.Bootstrapper{},
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

func TestPurchaseOrder_PackCoreDocument(t *testing.T) {
	id, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)

	po := new(PurchaseOrder)
	assert.NoError(t, po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), id.ID.String()))

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

func TestPO_UnpackCoreDocument(t *testing.T) {
	var model = new(PurchaseOrder)
	var err error

	// embed data missing
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{})
	assert.Error(t, err)

	// embed data type is wrong
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{EmbeddedData: new(any.Any)})
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// embed data is wrong
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{
		EmbeddedData: &any.Any{
			Value:   utils.RandomSlice(32),
			TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		},
	})
	assert.Error(t, err)

	// successful
	po, cd := createCDWithEmbeddedPO(t)
	err = model.UnpackCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, model.getClientData(), po.(*PurchaseOrder).getClientData())
	assert.Equal(t, model.ID(), po.ID())
	assert.Equal(t, model.CurrentVersion(), po.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), po.PreviousVersion())
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
	data.Recipient = "0xed03fa80291ff5ddc284de6b51e716b130b05e20"

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

	did, err := identity.NewDIDFromString("0xed03fa80291ff5ddc284de6b51e716b130b05e20")
	assert.NoError(t, err)
	assert.Equal(t, poModel.Recipient[:], did[:])
	assert.Equal(t, poModel.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
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

func TestPOModel_GenerateProofs(t *testing.T) {
	po := createPurchaseOrder(t)
	assert.NotNil(t, po)
	proof, err := po.CreateProofs([]string{"po.po_number", documents.CDTreePrefix + ".next_version", documents.CDTreePrefix + ".document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := po.DocumentRootTree()
	assert.NoError(t, err)

	// Validate po_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate next_version
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	assert.Equal(t, po.NextVersion(), proof[1].Value)

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestPOModel_createProofsFieldDoesNotExist(t *testing.T) {
	poModel := createPurchaseOrder(t)
	_, err := poModel.CreateProofs([]string{"nonexisting"})
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

func createPurchaseOrder(t *testing.T) *PurchaseOrder {
	po := new(PurchaseOrder)
	err := po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	_, err = po.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = po.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = po.CalculateDocumentRoot()
	assert.NoError(t, err)
	return po
}
