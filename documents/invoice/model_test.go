// +build unit

package invoice

import (
	"encoding/json"
	"reflect"
	"testing"

	"context"
	"os"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient

	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&queue.Bootstrapper{},
		&identity.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestInvoice_FromCoreDocuments_invalidParameter(t *testing.T) {
	invoiceModel := &Invoice{}

	emptyCoreDocument := &coredocumentpb.CoreDocument{}
	err := invoiceModel.UnpackCoreDocument(emptyCoreDocument)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	err = invoiceModel.UnpackCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	err = invoiceModel.UnpackCoreDocument(coreDocument)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestInvoice_InitCoreDocument_successful(t *testing.T) {
	invoiceModel := &Invoice{}

	coreDocument := testingdocuments.CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	err := invoiceModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}

func TestInvoice_InitCoreDocument_invalidCentId(t *testing.T) {
	invoiceModel := &Invoice{}

	coreDocument := testingdocuments.CreateCDWithEmbeddedInvoice(t, invoicepb.InvoiceData{
		Recipient:   utils.RandomSlice(identity.CentIDLength + 1),
		Sender:      utils.RandomSlice(identity.CentIDLength),
		Payee:       utils.RandomSlice(identity.CentIDLength),
		GrossAmount: 42,
	})
	err := invoiceModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err)
	assert.NotNil(t, invoiceModel.Sender)
	assert.NotNil(t, invoiceModel.Payee)
	assert.Nil(t, invoiceModel.Recipient)
}

func TestInvoice_CoreDocument_successful(t *testing.T) {
	invoiceModel := &Invoice{}

	//init model with a CoreDoc
	coreDocument := testingdocuments.CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	invoiceModel.UnpackCoreDocument(coreDocument)

	returnedCoreDocument, err := invoiceModel.PackCoreDocument()
	assert.Nil(t, err, "transformation from invoice to CoreDoc failed")

	assert.Equal(t, coreDocument.EmbeddedData, returnedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, returnedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")
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

	//init model with a CoreDoc
	coreDocument := testingdocuments.CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	invoiceModel.UnpackCoreDocument(coreDocument)

	jsonBytes, err := invoiceModel.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	err = invoiceModel.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	receivedCoreDocument, err := invoiceModel.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, receivedCoreDocument.EmbeddedData, coreDocument.EmbeddedData, "JSON unmarshal damaged invoice variables")
}

func TestInvoiceModel_UnpackCoreDocument(t *testing.T) {
	var model documents.Model = new(Invoice)
	var err error

	// nil core doc
	err = model.UnpackCoreDocument(nil)
	assert.Error(t, err, "unpack must fail")

	// embed data missing
	err = model.UnpackCoreDocument(new(coredocumentpb.CoreDocument))
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// successful
	coreDocument := testingdocuments.CreateCDWithEmbeddedInvoice(t, testingdocuments.CreateInvoiceData())
	err = model.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")

	receivedCoreDocument, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.Equal(t, coreDocument.EmbeddedData, receivedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, receivedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")
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
	contextHeader, err := header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)
	// fail recipient
	data := &clientinvoicepb.InvoiceData{
		Sender:    "some number",
		Payee:     "some payee",
		Recipient: "some recipient",
		ExtraData: "some data",
	}
	inv := new(Invoice)
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, contextHeader)
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "failed to decode extra data")
	assert.Nil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)
	assert.Nil(t, inv.ExtraData)

	data.ExtraData = "0x010203020301"
	data.Recipient = "0x010203040506"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, contextHeader)
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	data.Sender = "0x010203060506"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, contextHeader)
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	data.Payee = "0x010203030405"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, contextHeader)
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.NotNil(t, inv.Payee)

	data.ExtraData = "0x010203020301"
	collabs := []string{"0x010102040506", "some id"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs}, contextHeader)
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collabs = []string{"0x010102040506", "0x010203020302"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs}, contextHeader)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, inv.Sender[:], []byte{1, 2, 3, 6, 5, 6})
	assert.Equal(t, inv.Payee[:], []byte{1, 2, 3, 3, 4, 5})
	assert.Equal(t, inv.Recipient[:], []byte{1, 2, 3, 4, 5, 6})
	assert.Equal(t, inv.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
	id := contextHeader.Self().ID
	assert.Equal(t, inv.CoreDocument.Collaborators, [][]byte{id[:], {1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}})
}

func TestInvoiceModel_calculateDataRoot(t *testing.T) {
	ctxHeader, err := header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)
	m := new(Invoice)
	err = m.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), ctxHeader)
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, m.InvoiceSalts, "salts must be nil")

	err = m.calculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.NotNil(t, m.CoreDocument, "coredoc must be created")
	assert.NotNil(t, m.InvoiceSalts, "salts must be created")
	assert.NotNil(t, m.CoreDocument.DataRoot, "data root must be filled")
}

func TestInvoiceModel_createProofs(t *testing.T) {
	i, corDoc, err := createMockInvoice(t)
	assert.Nil(t, err)
	corDoc, proof, err := i.createProofs([]string{"invoice.invoice_number", "collaborators[0]", "document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	assert.NotNil(t, corDoc)
	tree, _ := coredocument.GetDocumentRootTree(corDoc)

	// Validate invoice_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate collaborators[0]
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate '0x' Hex format in []byte value
	assert.Equal(t, hexutil.Encode(i.CoreDocument.Collaborators[0]), proof[1].Value)

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i, _, err := createMockInvoice(t)
	assert.Nil(t, err)
	_, _, err = i.createProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestInvoiceModel_GetDocumentID(t *testing.T) {
	i, corDoc, err := createMockInvoice(t)
	assert.Nil(t, err)
	ID, err := i.ID()
	assert.Equal(t, corDoc.DocumentIdentifier, ID)
}

func TestInvoiceModel_getDocumentDataTree(t *testing.T) {
	i := Invoice{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2}
	tree, err := i.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("invoice.invoice_number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "invoice.invoice_number", leaf.Property)
}

func createMockInvoice(t *testing.T) (*Invoice, *coredocumentpb.CoreDocument, error) {
	i := &Invoice{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2, Currency: "USD", CoreDocument: coredocument.New()}
	i.CoreDocument.Collaborators = [][]byte{{1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}}
	err := i.calculateDataRoot()
	if err != nil {
		return nil, nil, err
	}
	// get the coreDoc for the invoice
	corDoc, err := i.PackCoreDocument()
	if err != nil {
		return nil, nil, err
	}
	assert.Nil(t, coredocument.FillSalts(corDoc))
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, nil, err
	}
	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, nil, err
	}
	i.UnpackCoreDocument(corDoc)
	return i, corDoc, nil
}
