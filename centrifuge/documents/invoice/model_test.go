// +build unit

package invoice

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestInvoice_FromCoreDocuments_invalidParameter(t *testing.T) {
	invoiceModel := &InvoiceModel{}

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
	invoiceModel := &InvoiceModel{}

	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
	err := invoiceModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}

func TestInvoice_InitCoreDocument_invalidCentId(t *testing.T) {
	invoiceModel := &InvoiceModel{}

	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, invoicepb.InvoiceData{
		Recipient:   tools.RandomSlice(identity.CentIDLength + 1),
		Sender:      tools.RandomSlice(identity.CentIDLength),
		Payee:       tools.RandomSlice(identity.CentIDLength),
		GrossAmount: 42,
	})
	err := invoiceModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err)
	assert.NotNil(t, invoiceModel.Sender)
	assert.NotNil(t, invoiceModel.Payee)
	assert.Nil(t, invoiceModel.Recipient)
}

func TestInvoice_CoreDocument_successful(t *testing.T) {
	invoiceModel := &InvoiceModel{}

	//init model with a CoreDoc
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
	invoiceModel.UnpackCoreDocument(coreDocument)

	returnedCoreDocument, err := invoiceModel.PackCoreDocument()
	assert.Nil(t, err, "transformation from invoice to CoreDoc failed")

	assert.Equal(t, coreDocument.EmbeddedData, returnedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, returnedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}

func TestInvoice_ModelInterface(t *testing.T) {
	var i interface{} = &InvoiceModel{}
	_, ok := i.(documents.Model)
	assert.True(t, ok, "model interface not implemented correctly for invoiceModel")
}

func TestInvoice_Type(t *testing.T) {
	var model documents.Model
	model = &InvoiceModel{}
	assert.Equal(t, model.Type(), reflect.TypeOf(&InvoiceModel{}), "InvoiceType not correct")
}

func TestInvoice_JSON(t *testing.T) {
	invoiceModel := &InvoiceModel{}

	//init model with a CoreDoc
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
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
	var model documents.Model = new(InvoiceModel)
	var err error

	// nil core doc
	err = model.UnpackCoreDocument(nil)
	assert.Error(t, err, "unpack must fail")

	// embed data missing
	err = model.UnpackCoreDocument(new(coredocumentpb.CoreDocument))
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// successful
	coreDocument := testinginvoice.CreateCDWithEmbeddedInvoice(t, testinginvoice.CreateInvoiceData())
	err = model.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid core document with embedded invoice shouldn't produce an error")

	receivedCoreDocument, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")

	assert.Equal(t, coreDocument.EmbeddedData, receivedCoreDocument.EmbeddedData, "embeddedData should be the same")
	assert.Equal(t, coreDocument.EmbeddedDataSalts, receivedCoreDocument.EmbeddedDataSalts, "embeddedDataSalt should be the same")
}

func TestInvoiceModel_getClientData(t *testing.T) {
	invData := testinginvoice.CreateInvoiceData()
	inv := new(InvoiceModel)
	inv.loadFromP2PProtobuf(&invData)

	data := inv.getClientData()
	assert.NotNil(t, data, "invoice data should not be nil")
	assert.Equal(t, data.GrossAmount, data.GrossAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, hexutil.Encode(inv.Recipient[:]), "recipient should match")
	assert.Equal(t, data.Sender, hexutil.Encode(inv.Sender[:]), "sender should match")
	assert.Equal(t, data.Payee, hexutil.Encode(inv.Payee[:]), "payee should match")
}

func TestInvoiceModel_InitInvoiceInput(t *testing.T) {
	// fail recipient
	data := &clientinvoicepb.InvoiceData{
		Sender:    "some number",
		Payee:     "some payee",
		Recipient: "some recipient",
		ExtraData: "some data",
	}
	inv := new(InvoiceModel)
	err := inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data})
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "failed to decode extra data")
	assert.Nil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)
	assert.Nil(t, inv.ExtraData)

	data.ExtraData = "0x010203020301"
	data.Recipient = "0x010203040506"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data})
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	data.Sender = "0x010203060506"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data})
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	data.Payee = "0x010203030405"
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data})
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.NotNil(t, inv.Payee)

	data.ExtraData = "0x010203020301"
	collabs := []string{"0x010102040506", "some id"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs})
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collabs = []string{"0x010102040506", "0x010203020302"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, Collaborators: collabs})
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, inv.Sender[:], []byte{1, 2, 3, 6, 5, 6})
	assert.Equal(t, inv.Payee[:], []byte{1, 2, 3, 3, 4, 5})
	assert.Equal(t, inv.Recipient[:], []byte{1, 2, 3, 4, 5, 6})
	assert.Equal(t, inv.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
	assert.Equal(t, inv.CoreDocument.Collaborators, [][]byte{{1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}})
}

func TestInvoiceModel_calculateDataRoot(t *testing.T) {
	m := new(InvoiceModel)
	err := m.InitInvoiceInput(createPayload())
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
	corDoc, proof, err := i.createProofs([]string{"invoice_number"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	assert.NotNil(t, corDoc)
	tree, _ := coredocument.GetDocumentRootTree(corDoc)
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i, _, err := createMockInvoice(t)
	assert.Nil(t, err)
	_, _, err = i.createProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestInvoiceModel_getDocumentDataTree(t *testing.T) {
	i := InvoiceModel{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2}
	tree, err := i.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("invoice_number")
	assert.Equal(t, "invoice_number", leaf.Property)
}

func createMockInvoice(t *testing.T) (*InvoiceModel, *coredocumentpb.CoreDocument, error) {
	i := &InvoiceModel{InvoiceNumber: "3213121", NetAmount: 2, GrossAmount: 2, CoreDocument: coredocument.New()}
	err := i.calculateDataRoot()
	if err != nil {
		return nil, nil, err
	}
	// get the coreDoc for the invoice
	corDoc, err := i.PackCoreDocument()
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
	i.UnpackCoreDocument(corDoc)
	return i, corDoc, nil
}
