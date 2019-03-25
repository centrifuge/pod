// +build unit

package purchaseorder

import (
	"encoding/json"
	"fmt"
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
	"github.com/centrifuge/go-centrifuge/errors"
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
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
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
var defaultDID = testingidentity.GenerateRandomDID()

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
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	po := new(PurchaseOrder)
	assert.NoError(t, po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), did.String()))

	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestPurchaseOrder_JSON(t *testing.T) {
	po := new(PurchaseOrder)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	assert.NoError(t, po.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), did.String()))

	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := po.JSON()
	assert.NoError(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	po = new(PurchaseOrder)
	err = po.FromJSON(jsonBytes)
	assert.NoError(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := po.PackCoreDocument()
	assert.NoError(t, err, "JSON unmarshal damaged invoice variables")
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
	err := poModel.loadFromP2PProtobuf(&poData)
	assert.NoError(t, err)

	data := poModel.getClientData()
	assert.NotNil(t, data, "purchase order data should not be nil")
	assert.Equal(t, data.TotalAmount, data.TotalAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, poModel.Recipient.String(), "recipient should match")
}

func TestPOOrderModel_InitPOInput(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	// fail recipient
	data := &clientpurchaseorderpb.PurchaseOrderData{
		Recipient: "some recipient",
	}
	poModel := new(PurchaseOrder)
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data}, did.String())
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "malformed address provided")
	assert.Nil(t, poModel.Recipient)

	data.Recipient = "0xed03fa80291ff5ddc284de6b51e716b130b05e20"
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data}, did.String())
	assert.Nil(t, err)
	assert.NotNil(t, poModel.Recipient)

	collabs := []string{"0x010102040506", "some id"}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs}, did.String())
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collab1, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	collab2, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3")
	assert.NoError(t, err)
	collabs = []string{collab1.String(), collab2.String()}
	err = poModel.InitPurchaseOrderInput(&clientpurchaseorderpb.PurchaseOrderCreatePayload{Data: data, Collaborators: collabs}, did.String())
	assert.Nil(t, err, "must be nil")

	did, err = identity.NewDIDFromString("0xed03fa80291ff5ddc284de6b51e716b130b05e20")
	assert.NoError(t, err)
	assert.Equal(t, poModel.Recipient[:], did[:])
}

func TestPOModel_calculateDataRoot(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	poModel := new(PurchaseOrder)
	err = poModel.InitPurchaseOrderInput(testingdocuments.CreatePOPayload(), did.String())
	assert.Nil(t, err, "Init must pass")

	dr, err := poModel.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
}

func TestPOModel_CreateProofs(t *testing.T) {
	po := createPurchaseOrder(t)
	assert.NotNil(t, po)
	rk := po.CoreDocument.GetTestCoreDocWithReset().Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := po.CreateProofs([]string{"po.number", pf, documents.CDTreePrefix + ".document_type", "po.line_items[0].status"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := po.DocumentRootTree()
	assert.NoError(t, err)

	// Validate po_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles collaborators
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc := identity.NewDIDFromBytes(proof[1].Value)
	assert.True(t, po.AccountCanRead(acc))

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)

	// validate line items
	valid, err = tree.ValidateProof(proof[3])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestPOModel_createProofsFieldDoesNotExist(t *testing.T) {
	poModel := createPurchaseOrder(t)
	_, err := poModel.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestPOModel_getDocumentDataTree(t *testing.T) {
	na := new(documents.Decimal)
	assert.NoError(t, na.SetString("2"))
	poModel := createPurchaseOrder(t)
	poModel.Number = "123"
	poModel.TotalAmount = na
	tree, err := poModel.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("po.number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "po.number", leaf.Property.ReadableName())
	assert.Equal(t, []byte(poModel.Number), leaf.Value)
}

func createPurchaseOrder(t *testing.T) *PurchaseOrder {
	po := new(PurchaseOrder)
	payload := testingdocuments.CreatePOPayload()
	payload.Data.LineItems = []*clientpurchaseorderpb.LineItem{
		{
			Status:      "pending",
			AmountTotal: "1.1",
			Activities: []*clientpurchaseorderpb.LineItemActivity{
				{
					ItemNumber: "12345",
					Status:     "pending",
					Amount:     "1.1",
				},
			},
			TaxItems: []*clientpurchaseorderpb.TaxItem{
				{
					ItemNumber: "12345",
					TaxAmount:  "1.1",
				},
			},
		},
	}
	err := po.InitPurchaseOrderInput(payload, defaultDID.String())
	assert.NoError(t, err)
	po.GetTestCoreDocWithReset()
	_, err = po.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = po.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = po.CalculateDocumentRoot()
	assert.NoError(t, err)
	return po
}

func TestPurchaseOrder_CollaboratorCanUpdate(t *testing.T) {
	po := createPurchaseOrder(t)
	id1 := defaultDID
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()

	// wrong type
	err := po.CollaboratorCanUpdate(new(mockModel), id1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))
	assert.NoError(t, testRepo().Create(id1[:], po.CurrentVersion(), po))

	// update the document
	model, err := testRepo().Get(id1[:], po.CurrentVersion())
	assert.NoError(t, err)
	oldPO := model.(*PurchaseOrder)
	data := oldPO.getClientData()
	data.TotalAmount = "50"
	err = po.PrepareNewVersion(po, data, []string{id3.String()})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldPO.CollaboratorCanUpdate(po, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldPO.CollaboratorCanUpdate(po, id2))

	// update the id3 rules to update only total amount
	po.CoreDocument.Document.TransitionRules[3].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT
	po.CoreDocument.Document.TransitionRules[3].Field = append(compactPrefix(), 0, 0, 0, 18)
	po.CoreDocument.Document.DocumentRoot = utils.RandomSlice(32)
	assert.NoError(t, testRepo().Create(id1[:], po.CurrentVersion(), po))

	// fetch the document
	model, err = testRepo().Get(id1[:], po.CurrentVersion())
	assert.NoError(t, err)
	oldPO = model.(*PurchaseOrder)
	data = oldPO.getClientData()
	data.TotalAmount = "55"
	data.Currency = "INR"
	err = po.PrepareNewVersion(po, data, nil)
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldPO.CollaboratorCanUpdate(po, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldPO.CollaboratorCanUpdate(po, id2))

	// id3 should fail with just one error since changing Currency is not allowed
	err = oldPO.CollaboratorCanUpdate(po, id3)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))
	assert.Contains(t, err.Error(), "po.currency")
}
