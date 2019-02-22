// +build unit

package invoice

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ethid"
	"github.com/centrifuge/go-centrifuge/p2p"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
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

	ibootstrappers := []bootstrap.TestBootstrapper{
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
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", cid.String())
	configService = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestInvoice_PackCoreDocument(t *testing.T) {
	id, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)

	inv := new(Invoice)
	assert.Error(t, inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), id.ID.String()))

	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
	assert.NotNil(t, cd.EmbeddedDataSalts)
}

func TestInvoice_JSON(t *testing.T) {
	inv := new(Invoice)
	id, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)
	assert.NoError(t, inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), id.ID.String()))

	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := inv.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	inv = new(Invoice)
	err = inv.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := inv.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, cd, ncd)
}

// TODO validate Unpack and remove if required
//func TestInvoiceModel_UnpackCoreDocument(t *testing.T) {
//	var model = new(Invoice)
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
//	coreDocumentModel := CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
//	model.CoreDocumentModel = coreDocumentModel
//	err = model.UnpackCoreDocument(coreDocumentModel)
//	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")
//
//	receivedCoreDocumentModel, err := model.PackCoreDocument()
//	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")
//
//	assert.Equal(t, coreDocumentModel.Document.EmbeddedData, receivedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be the same")
//	assert.Equal(t, coreDocumentModel.Document.EmbeddedDataSalts, receivedCoreDocumentModel.Document.EmbeddedDataSalts, "embeddedDataSalt should be the same")
//}

func TestInvoiceModel_getClientData(t *testing.T) {
	invData := testingdocuments.CreateInvoiceData()
	inv := new(Invoice)
	inv.loadFromP2PProtobuf(&invData)

	data := inv.getClientData()
	assert.NotNil(t, data, "invoice data should not be nil")
	assert.Equal(t, data.GrossAmount, data.GrossAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, hexutil.Encode(inv.Recipient[:]), "recipient should match")
	assert.Equal(t, data.Sender, hexutil.Encode(inv.Sender[:]), "sender should match")
	assert.Equal(t, data.Payee, hexutil.Encode(inv.Payee[:]), "payee should match")
}

func TestInvoiceModel_InitInvoiceInput(t *testing.T) {
	id, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	// fail recipient
	data := &clientinvoicepb.InvoiceData{
		Sender:    "some number",
		Payee:     "some payee",
		Recipient: "some recipient",
		ExtraData: "some data",
	}
	inv := new(Invoice)
	err := inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, id.ID.String())
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "failed to decode extra data")
	assert.Nil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)
	assert.Nil(t, inv.ExtraData)

	data.ExtraData = "0x010203020301"
	data.Recipient = "0x010203040506"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, id.ID.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	data.Sender = "0x010203060506"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, id.ID.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	data.Payee = "0x010203030405"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, id.ID.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.NotNil(t, inv.Payee)

	data.ExtraData = "0x010203020301"
	collabs := []string{"0x010102040506", "some id"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs}, id.ID.String())
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collabs = []string{"0x010102040506", "0x010203020302"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs}, id.ID.String())
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, inv.Sender[:], []byte{1, 2, 3, 6, 5, 6})
	assert.Equal(t, inv.Payee[:], []byte{1, 2, 3, 3, 4, 5})
	assert.Equal(t, inv.Recipient[:], []byte{1, 2, 3, 4, 5, 6})
	assert.Equal(t, inv.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
}

func TestInvoiceModel_calculateDataRoot(t *testing.T) {
	id, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	m := new(Invoice)
	err := m.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), id.ID.String())
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, m.InvoiceSalts, "salts must be nil")

	dr, err := m.DataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
	assert.NotNil(t, m.InvoiceSalts, "salts must be created")
}

func TestInvoice_GenerateProofs(t *testing.T) {
	i, err := createInvoice(t)
	assert.Nil(t, err)
	proof, err := i.GenerateProofs([]string{"invoice.invoice_number", "collaborators[0]", "document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := i.CoreDocument.DocumentRootTree()
	assert.NoError(t, err)

	// Validate invoice_number
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
	assert.True(t, i.CoreDocument.AccountCanRead(id))

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i, err := createInvoice(t)
	assert.Nil(t, err)
	_, err = i.GenerateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestInvoiceModel_GetDocumentID(t *testing.T) {
	i, err := createInvoice(t)
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.ID(), i.ID())
}

func TestInvoiceModel_getDocumentDataTree(t *testing.T) {
	i := Invoice{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2}
	tree, err := i.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("invoice.invoice_number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "invoice.invoice_number", leaf.Property.ReadableName())
}

func createInvoice(t *testing.T) (*Invoice, error) {
	i := new(Invoice)
	i.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), "0x010203040506")
	_, err := i.DataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)
	return i, nil
}
