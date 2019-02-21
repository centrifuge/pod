// +build unit

package invoice

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"

	"github.com/centrifuge/go-centrifuge/identity/ideth"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/testingtx"
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

func TestInvoice_FromCoreDocuments_invalidParameter(t *testing.T) {
	invoiceModel := &Invoice{}

	emptyCoreDocModel := &documents.CoreDocumentModel{
		nil,
		nil,
	}
	err := invoiceModel.UnpackCoreDocument(emptyCoreDocModel)
	assert.Error(t, err, "it should not be possible to init with an empty core document")

	err = invoiceModel.UnpackCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init with an empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	coreDocModel := &documents.CoreDocumentModel{
		coreDocument,
		nil,
	}
	err = invoiceModel.UnpackCoreDocument(coreDocModel)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestInvoice_InitCoreDocument_successful(t *testing.T) {
	invoiceModel := &Invoice{}

	dm := CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	invoiceModel.CoreDocumentModel = dm
	err := invoiceModel.UnpackCoreDocument(dm)
	assert.Nil(t, err, "valid coredocumentmodel shouldn't produce an error")
}

func TestInvoice_InitCoreDocument_invalidCentId(t *testing.T) {
	invoiceModel := &Invoice{}

	dm := CreateCDWithEmbeddedInvoice(t, invoicepb.InvoiceData{
		Recipient:   utils.RandomSlice(identity.CentIDLength + 1),
		Sender:      utils.RandomSlice(identity.CentIDLength),
		Payee:       utils.RandomSlice(identity.CentIDLength),
		GrossAmount: 42,
	})
	invoiceModel.CoreDocumentModel = dm
	err := invoiceModel.UnpackCoreDocument(dm)
	assert.Nil(t, err)
	assert.NotNil(t, invoiceModel.Sender)
	assert.NotNil(t, invoiceModel.Payee)
	assert.Nil(t, invoiceModel.Recipient)
}

func TestInvoice_CoreDocument_successful(t *testing.T) {
	invoiceModel := &Invoice{}

	//init model with a CoreDocModel

	coreDocumentModel := CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	invoiceModel.CoreDocumentModel = coreDocumentModel
	invoiceModel.UnpackCoreDocument(coreDocumentModel)

	returnedCoreDocumentModel, err := invoiceModel.PackCoreDocument()
	assert.Nil(t, err, "transformation from invoice to CoreDocModel failed")

	assert.Equal(t, coreDocumentModel.Document.EmbeddedData, returnedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocumentModel.Document.EmbeddedDataSalts, returnedCoreDocumentModel.Document.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}

func TestInvoice_ModelInterface(t *testing.T) {
	var i interface{} = &Invoice{}
	_, ok := i.(documents.Model)
	assert.True(t, ok, "model interface not implemented correctly for invoiceModel")
}

func TestInvoice_Type(t *testing.T) {
	var model documents.Model
	model = &Invoice{}
	assert.Equal(t, model.Type(), reflect.TypeOf(&Invoice{}), "InvoiceType not correct")
}

func TestInvoice_JSON(t *testing.T) {
	invoiceModel := &Invoice{}

	//init model with a CoreDocModel
	coreDocumentModel := CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	invoiceModel.CoreDocumentModel = coreDocumentModel
	invoiceModel.UnpackCoreDocument(coreDocumentModel)

	jsonBytes, err := invoiceModel.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	err = invoiceModel.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	receivedCoreDocumentModel, err := invoiceModel.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, receivedCoreDocumentModel.Document.EmbeddedData, coreDocumentModel.Document.EmbeddedData, "JSON unmarshal damaged invoice variables")
}

func TestInvoiceModel_UnpackCoreDocument(t *testing.T) {
	var model = new(Invoice)
	var err error

	// nil core doc
	err = model.UnpackCoreDocument(nil)
	assert.Error(t, err, "unpack must fail")

	// embed data missing
	err = model.UnpackCoreDocument(new(documents.CoreDocumentModel))
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// successful
	coreDocumentModel := CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	model.CoreDocumentModel = coreDocumentModel
	err = model.UnpackCoreDocument(coreDocumentModel)
	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")

	receivedCoreDocumentModel, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.Equal(t, coreDocumentModel.Document.EmbeddedData, receivedCoreDocumentModel.Document.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocumentModel.Document.EmbeddedDataSalts, receivedCoreDocumentModel.Document.EmbeddedDataSalts, "embeddedDataSalt should be the same")
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
	assert.Equal(t, inv.CoreDocumentModel.Document.Collaborators, [][]byte{id.ID[:], collab1[:], collab2[:]})
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

func TestInvoiceModel_createProofs(t *testing.T) {
	i, err := createMockInvoice(t)
	assert.Nil(t, err)
	proof, err := i.CreateProofs([]string{"invoice.invoice_number", "collaborators[0]", "document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := i.CoreDocumentModel.GetDocumentRootTree()
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
	assert.Equal(t, i.CoreDocumentModel.Document.Collaborators[0], proof[1].Value)

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i, err := createMockInvoice(t)
	assert.Nil(t, err)
	_, err = i.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestInvoiceModel_GetDocumentID(t *testing.T) {
	i, err := createMockInvoice(t)
	assert.Nil(t, err)
	ID, err := i.ID()
	assert.Equal(t, i.CoreDocumentModel.Document.DocumentIdentifier, ID)
}

func TestInvoiceModel_getDocumentDataTree(t *testing.T) {
	i := Invoice{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2}
	tree, err := i.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("invoice.invoice_number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "invoice.invoice_number", leaf.Property.ReadableName())
}

func createMockInvoice(t *testing.T) (*Invoice, error) {
	i := &Invoice{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2, Currency: "USD", CoreDocumentModel: documents.NewCoreDocModel()}
	i.CoreDocumentModel.Document.Collaborators = [][]byte{{1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}}
	dataRoot, err := i.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the invoice

	cdm, err := i.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	err = cdm.CalculateSigningRoot(dataRoot)
	if err != nil {
		return nil, err
	}
	err = cdm.CalculateDocumentRoot()
	if err != nil {
		return nil, err
	}
	i.UnpackCoreDocument(cdm)
	return i, nil
}
