// +build unit

package invoice

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
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
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

func TestInvoice_PackCoreDocument(t *testing.T) {
	id, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)

	inv := new(Invoice)
	assert.NoError(t, inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), id.ID.String()))

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

func TestInvoiceModel_UnpackCoreDocument(t *testing.T) {
	var model = new(Invoice)
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
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
		},
	})
	assert.Error(t, err)

	// successful
	inv, cd := createCDWithEmbeddedInvoice(t)
	err = model.UnpackCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, model.getClientData(), inv.(*Invoice).getClientData())
	assert.Equal(t, model.ID(), inv.ID())
	assert.Equal(t, model.CurrentVersion(), inv.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), inv.PreviousVersion())
}

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
	recipientDID := testingidentity.GenerateRandomDID()
	data.Recipient = recipientDID.String()
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, id.ID.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	senderDID := testingidentity.GenerateRandomDID()
	data.Sender = senderDID.String()
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, id.ID.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	payeeDID := testingidentity.GenerateRandomDID()
	data.Payee = payeeDID.String()
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

	collab1, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	collab2, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3")
	assert.NoError(t, err)
	collabs = []string{collab1.String(), collab2.String()}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs}, id.ID.String())
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, inv.Sender[:], senderDID[:])
	assert.Equal(t, inv.Payee[:], payeeDID[:])
	assert.Equal(t, inv.Recipient[:], recipientDID[:])
	assert.Equal(t, inv.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
}

func TestInvoiceModel_calculateDataRoot(t *testing.T) {
	id, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	m := new(Invoice)
	err := m.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), id.ID.String())
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, m.InvoiceSalts, "salts must be nil")

	dr, err := m.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
	assert.NotNil(t, m.InvoiceSalts, "salts must be created")
}

func TestInvoice_GenerateProofs(t *testing.T) {
	i, err := createInvoice(t)
	assert.Nil(t, err)
	proof, err := i.CreateProofs([]string{"invoice.invoice_number", documents.CDTreePrefix + ".next_version", documents.CDTreePrefix + ".document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := i.CoreDocument.DocumentRootTree()
	assert.NoError(t, err)

	// Validate invoice_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate next_version
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	assert.Equal(t, i.NextVersion(), proof[1].Value)

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i, err := createInvoice(t)
	assert.Nil(t, err)
	_, err = i.CreateProofs([]string{"nonexisting"})
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
	err := i.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	_, err = i.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)
	return i, nil
}
