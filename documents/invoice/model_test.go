// +build unit

package invoice

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

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
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var configService config.Service
var defaultDID = testingidentity.GenerateRandomDID()

func TestMain(m *testing.M) {
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	jobMan := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobMan
	done := make(chan bool)
	jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), done, nil)
	ctx[bootstrap.BootstrappedInvoiceUnpaid] = new(testingdocuments.MockRegistry)
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
	cfg.Set("identityId", did.String())
	configService = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestInvoice_PackCoreDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	inv := new(Invoice)
	assert.NoError(t, inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), did))

	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestInvoice_JSON(t *testing.T) {
	inv := new(Invoice)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	assert.NoError(t, inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), did))

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
	d, err := model.getClientData()
	assert.NoError(t, err)
	d1, err := inv.(*Invoice).getClientData()
	assert.NoError(t, err)
	assert.Equal(t, d, d1)
	assert.Equal(t, model.ID(), inv.ID())
	assert.Equal(t, model.CurrentVersion(), inv.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), inv.PreviousVersion())
}

func TestInvoiceModel_getClientData(t *testing.T) {
	invData := testingdocuments.CreateInvoiceData()
	inv := new(Invoice)
	inv.CoreDocument = new(documents.CoreDocument)
	err := inv.loadFromP2PProtobuf(&invData)
	assert.NoError(t, err)

	data, err := inv.getClientData()
	assert.NoError(t, err)
	assert.NotNil(t, data, "invoice data should not be nil")
	assert.Equal(t, data.GrossAmount, data.GrossAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, inv.Data.Recipient.String(), "recipient should match")
	assert.Equal(t, data.Sender, inv.Data.Sender.String(), "sender should match")
	assert.Equal(t, data.Payee, inv.Data.Payee.String(), "payee should match")
}

func TestInvoiceModel_InitInvoiceInput(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	// fail recipient
	data := &clientinvoicepb.InvoiceData{
		Recipient: "some recipient",
	}
	inv := new(Invoice)
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, did)
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "malformed address provided")
	assert.Nil(t, inv.Data.Recipient)
	assert.Nil(t, inv.Data.Sender)
	assert.Nil(t, inv.Data.Payee)

	recipientDID := testingidentity.GenerateRandomDID()
	data.Recipient = recipientDID.String()
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, did)
	assert.Nil(t, err)
	assert.NotNil(t, inv.Data.Recipient)
	assert.Nil(t, inv.Data.Sender)
	assert.Nil(t, inv.Data.Payee)

	senderDID := testingidentity.GenerateRandomDID()
	data.Sender = senderDID.String()
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, did)
	assert.Nil(t, err)
	assert.NotNil(t, inv.Data.Recipient)
	assert.NotNil(t, inv.Data.Sender)
	assert.Nil(t, inv.Data.Payee)

	payeeDID := testingidentity.GenerateRandomDID()
	data.Payee = payeeDID.String()
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data}, did)
	assert.Nil(t, err)
	assert.NotNil(t, inv.Data.Recipient)
	assert.NotNil(t, inv.Data.Sender)
	assert.NotNil(t, inv.Data.Payee)

	collabs := []string{"0x010102040506", "some id"}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, WriteAccess: collabs}, did)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "malformed address provided")

	collab1, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	collab2, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3")
	assert.NoError(t, err)
	collabs = []string{collab1.String(), collab2.String()}
	err = inv.InitInvoiceInput(&clientinvoicepb.InvoiceCreatePayload{Data: data, WriteAccess: collabs}, did)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, inv.Data.Sender[:], senderDID[:])
	assert.Equal(t, inv.Data.Payee[:], payeeDID[:])
	assert.Equal(t, inv.Data.Recipient[:], recipientDID[:])
}

func TestInvoiceModel_calculateDataRoot(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	m := new(Invoice)
	err = m.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), did)
	assert.Nil(t, err, "Init must pass")

	dr, err := m.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
}

func TestInvoice_CreateProofs(t *testing.T) {
	i := createInvoice(t)
	rk := i.GetTestCoreDocWithReset().Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := i.CreateProofs([]string{"invoice.number", pf, documents.CDTreePrefix + ".document_type", "invoice.line_items[0].item_number", "invoice.line_items[0].description"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	if err != nil {
		return
	}

	tree, err := i.DocumentRootTree()
	assert.NoError(t, err)

	// Validate invoice_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := identity.NewDIDFromBytes(proof[1].Value)
	assert.NoError(t, err)
	assert.True(t, i.AccountCanRead(acc))

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)

	// validate line item
	valid, err = tree.ValidateProof(proof[3])
	assert.Nil(t, err)
	assert.True(t, valid)

	valid, err = tree.ValidateProof(proof[4])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoice_CreateNFTProofs(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	acc := tc.(*configstore.Account)
	acc.IdentityID = defaultDID[:]
	assert.NoError(t, err)
	i := new(Invoice)
	invPayload := testingdocuments.CreateInvoicePayload()
	invPayload.Data.DateDue = &timestamp.Timestamp{Seconds: time.Now().Unix()}
	invPayload.Data.Status = "unpaid"
	invPayload.WriteAccess = []string{defaultDID.String()}
	err = i.InitInvoiceInput(invPayload, defaultDID)
	assert.NoError(t, err)
	sig, err := acc.SignMsg([]byte{0, 1, 2, 3})
	assert.NoError(t, err)
	i.AppendSignatures(sig)
	_, err = i.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)

	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(defaultDID[:], keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerId)
	proofFields := []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due", "invoice.sender", "invoice.status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"}
	proof, err := i.CreateProofs(proofFields)
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := i.DocumentRootTree()
	assert.NoError(t, err)
	assert.Len(t, proofFields, 8)

	// Validate invoice_gross_amount
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate signing_root
	valid, err = tree.ValidateProof(proof[5])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate signature
	valid, err = tree.ValidateProof(proof[6])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate next_version
	valid, err = tree.ValidateProof(proof[7])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i := createInvoice(t)
	_, err := i.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestInvoiceModel_GetDocumentID(t *testing.T) {
	i := createInvoice(t)
	assert.Equal(t, i.CoreDocument.ID(), i.ID())
}

func TestInvoiceModel_getDocumentDataTree(t *testing.T) {
	na := new(documents.Decimal)
	assert.NoError(t, na.SetString("2"))
	ga := new(documents.Decimal)
	assert.NoError(t, ga.SetString("2"))
	i := createInvoice(t)
	i.Data.Number = "321321"
	i.Data.NetAmount = na
	i.Data.GrossAmount = ga
	tree, err := i.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("invoice.number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "invoice.number", leaf.Property.ReadableName())
	assert.Equal(t, []byte(i.Data.Number), leaf.Value)
}

func createInvoice(t *testing.T) *Invoice {
	i := new(Invoice)
	payload := testingdocuments.CreateInvoicePayload()
	payload.Data.LineItems = []*clientinvoicepb.LineItem{
		{
			ItemNumber:  "123456",
			TaxAmount:   "1.99",
			TotalAmount: "99",
			Description: "Some description",
		},
	}

	err := i.InitInvoiceInput(payload, defaultDID)
	assert.NoError(t, err)
	i.GetTestCoreDocWithReset()
	_, err = i.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)
	return i
}

func TestInvoice_CollaboratorCanUpdate(t *testing.T) {
	inv := createInvoice(t)
	id1 := defaultDID
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()

	// wrong type
	err := inv.CollaboratorCanUpdate(new(mockModel), id1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))
	assert.NoError(t, testRepo().Create(id1[:], inv.CurrentVersion(), inv))

	// update the document
	model, err := testRepo().Get(id1[:], inv.CurrentVersion())
	assert.NoError(t, err)
	oldInv := model.(*Invoice)
	data, err := oldInv.getClientData()
	assert.NoError(t, err)
	data.GrossAmount = "50"
	err = inv.PrepareNewVersion(inv, data, documents.CollaboratorsAccess{
		ReadWriteCollaborators: []identity.DID{id3},
	})
	assert.NoError(t, err)

	_, err = inv.CalculateDataRoot()
	assert.NoError(t, err)

	_, err = inv.CalculateSigningRoot()
	assert.NoError(t, err)

	_, err = inv.CalculateDocumentRoot()
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldInv.CollaboratorCanUpdate(inv, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldInv.CollaboratorCanUpdate(inv, id2))

	// update the id3 rules to update only gross amount
	inv.CoreDocument.GetTestCoreDocWithReset().TransitionRules[3].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT
	inv.CoreDocument.GetTestCoreDocWithReset().TransitionRules[3].Field = append(compactPrefix(), 0, 0, 0, 14)
	assert.NoError(t, testRepo().Create(id1[:], inv.CurrentVersion(), inv))

	// fetch the document
	model, err = testRepo().Get(id1[:], inv.CurrentVersion())
	assert.NoError(t, err)
	oldInv = model.(*Invoice)
	data, err = oldInv.getClientData()
	assert.NoError(t, err)
	data.GrossAmount = "55"
	data.Currency = "INR"
	err = inv.PrepareNewVersion(inv, data, documents.CollaboratorsAccess{})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldInv.CollaboratorCanUpdate(inv, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldInv.CollaboratorCanUpdate(inv, id2))

	// id3 should fail with just one error since changing Currency is not allowed
	err = oldInv.CollaboratorCanUpdate(inv, id3)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))
	assert.Contains(t, err.Error(), "invoice.currency")
}

func TestInvoice_AddAttributes(t *testing.T) {
	inv, _ := createCDWithEmbeddedInvoice(t)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// success
	err = inv.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.NoError(t, err)
	assert.True(t, inv.AttributeExists(attr.Key))
	gattr, err := inv.GetAttribute(attr.Key)
	assert.NoError(t, err)
	assert.Equal(t, attr, gattr)

	// fail
	attr.Value.Type = documents.AttributeType("some attr")
	err = inv.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))
}

func TestInvoice_DeleteAttribute(t *testing.T) {
	inv, _ := createCDWithEmbeddedInvoice(t)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// failed
	err = inv.DeleteAttribute(attr.Key, true)
	assert.Error(t, err)

	// success
	assert.NoError(t, inv.AddAttributes(documents.CollaboratorsAccess{}, true, attr))
	assert.True(t, inv.AttributeExists(attr.Key))
	assert.NoError(t, inv.DeleteAttribute(attr.Key, true))
	assert.False(t, inv.AttributeExists(attr.Key))
}

func TestInvoice_GetData(t *testing.T) {
	inv := createInvoice(t)
	data := inv.GetData()
	assert.Equal(t, inv.Data, data)
}

func marshallData(t *testing.T, m map[string]interface{}) []byte {
	data, err := json.Marshal(m)
	assert.NoError(t, err)
	return data
}

func emptyDecimalData(t *testing.T) []byte {
	d := map[string]interface{}{
		"gross_amount": "",
	}

	return marshallData(t, d)
}

func invalidDecimalData(t *testing.T) []byte {
	d := map[string]interface{}{
		"gross_amount": "10.10.",
	}

	return marshallData(t, d)
}

func emptyDIDData(t *testing.T) []byte {
	d := map[string]interface{}{
		"recipient": "",
	}

	return marshallData(t, d)
}

func invalidDIDData(t *testing.T) []byte {
	d := map[string]interface{}{
		"recipient": "1acdew123asdefres",
	}

	return marshallData(t, d)
}

func emptyTimeData(t *testing.T) []byte {
	d := map[string]interface{}{
		"date_due": "",
	}

	return marshallData(t, d)
}

func invalidTimeData(t *testing.T) []byte {
	d := map[string]interface{}{
		"date_due": "1920-12-10",
	}

	return marshallData(t, d)
}

func validData(t *testing.T) []byte {
	d := map[string]interface{}{
		"number":       "12345",
		"status":       "unpaid",
		"gross_amount": "12.345",
		"recipient":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_paid":    "2019-05-24T14:48:44Z",        // rfc3339
		"attachments": []map[string]interface{}{
			{
				"name":      "test",
				"file_type": "pdf",
				"size":      1000202,
				"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
			},
		},
	}

	return marshallData(t, d)
}

func validDataWithCurrency(t *testing.T) []byte {
	d := map[string]interface{}{
		"number":       "12345",
		"status":       "unpaid",
		"gross_amount": "12.345",
		"recipient":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_paid":    "2019-05-24T14:48:44Z",        // rfc3339
		"currency":     "EUR",
		"attachments": []map[string]interface{}{
			{
				"name":      "test",
				"file_type": "pdf",
				"size":      1000202,
				"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
			},
		},
	}

	return marshallData(t, d)
}

func checkInvoicePayloadDataError(t *testing.T, inv *Invoice, payload documents.CreatePayload) {
	err := inv.loadData(payload.Data)
	assert.Error(t, err)
}

func TestInvoice_loadData(t *testing.T) {
	inv := new(Invoice)
	payload := documents.CreatePayload{}

	// empty decimal data
	payload.Data = emptyDecimalData(t)
	checkInvoicePayloadDataError(t, inv, payload)

	// invalid decimal data
	payload.Data = invalidDecimalData(t)
	checkInvoicePayloadDataError(t, inv, payload)

	// empty did data
	payload.Data = emptyDIDData(t)
	checkInvoicePayloadDataError(t, inv, payload)

	// invalid did data
	payload.Data = invalidDIDData(t)
	checkInvoicePayloadDataError(t, inv, payload)

	// empty time data
	payload.Data = emptyTimeData(t)
	checkInvoicePayloadDataError(t, inv, payload)

	// invalid time data
	payload.Data = invalidTimeData(t)
	checkInvoicePayloadDataError(t, inv, payload)

	// valid data
	payload.Data = validData(t)
	err := inv.loadData(payload.Data)
	assert.NoError(t, err)
	data := inv.GetData().(Data)
	assert.Equal(t, data.Number, "12345")
	assert.Equal(t, data.Status, "unpaid")
	assert.Equal(t, data.GrossAmount.String(), "12.345")
	assert.Equal(t, data.Recipient.String(), "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.Equal(t, data.DateDue.UTC().Format(time.RFC3339Nano), "2019-05-24T14:48:44.308854Z")
	assert.Equal(t, data.DatePaid.UTC().Format(time.RFC3339), "2019-05-24T14:48:44Z")
	assert.Len(t, data.Attachments, 1)
	assert.Equal(t, data.Attachments[0].Name, "test")
	assert.Equal(t, data.Attachments[0].FileType, "pdf")
	assert.Equal(t, data.Attachments[0].Size, 1000202)
	assert.Equal(t, hexutil.Encode(data.Attachments[0].Checksum), "0xbaeb33a61f05e6f269f1c4b4cff91a901b54daf3")
	assert.Equal(t, hexutil.Encode(data.Attachments[0].Data), "0xbaeb33a61f05e6f269f1c4b4cff91a901b54daf7")
}

func TestInvoice_unpackFromCreatePayload(t *testing.T) {
	payload := documents.CreatePayload{}
	inv := new(Invoice)

	// invalid data
	payload.Data = invalidDecimalData(t)
	err := inv.unpackFromCreatePayload(did, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvoiceInvalidData, err))

	// invalid attributes
	attr, err := documents.NewAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttributeType("some type")
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	payload.Data = validData(t)
	err = inv.unpackFromCreatePayload(did, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = inv.unpackFromCreatePayload(did, payload)
	assert.NoError(t, err)
}

func TestInvoice_unpackFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	old := createInvoice(t)
	inv := new(Invoice)

	// invalid data
	payload.Data = invalidDecimalData(t)
	err := inv.unpackFromUpdatePayload(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvoiceInvalidData, err))

	// invalid attributes
	attr, err := documents.NewAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttributeType("some type")
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	payload.Data = validData(t)
	err = inv.unpackFromUpdatePayload(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = inv.unpackFromUpdatePayload(old, payload)
	assert.NoError(t, err)
}
