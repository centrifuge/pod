// +build unit

package invoice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/precise-proofs/proofs"

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
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang.org/x/crypto/blake2b"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var defaultDID = testingidentity.GenerateRandomDID()

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func TestMain(m *testing.M) {
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	jobMan := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobMan
	done := make(chan error)
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
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestInvoice_PackCoreDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	inv := new(Invoice)
	assert.NoError(t, inv.DeriveFromCreatePayload(ctx, CreateInvoicePayload(t, []identity.DID{did})))
	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestInvoice_JSON(t *testing.T) {
	inv := new(Invoice)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	assert.NoError(t, inv.DeriveFromCreatePayload(ctx, CreateInvoicePayload(t, []identity.DID{did})))

	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := inv.JSON()
	assert.NoError(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	inv = new(Invoice)
	err = inv.FromJSON(jsonBytes)
	assert.NoError(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := inv.PackCoreDocument()
	assert.NoError(t, err, "JSON unmarshal damaged invoice variables")
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
			TypeUrl: documenttypes.EntityDataTypeUrl,
		},
	})
	assert.Error(t, err)

	// successful
	inv, cd := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	assert.NoError(t, model.UnpackCoreDocument(cd))
	data := model.GetData()
	data1 := inv.GetData()
	assert.Equal(t, data, data1)
	assert.Equal(t, model.ID(), inv.ID())
	assert.Equal(t, model.CurrentVersion(), inv.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), inv.PreviousVersion())
}

func TestInvoice_CreateProofs(t *testing.T) {
	i, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	rk := i.GetTestCoreDocWithReset().Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := i.CreateProofs([]string{"invoice.number", pf, documents.CDTreePrefix + ".document_type", "invoice.line_items[0].item_number", "invoice.line_items[0].description"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	if err != nil {
		return
	}

	dataRoot := calculateBasicDataRoot(t, i)

	nodeHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate invoice_number
	valid, err := documents.ValidateProof(proof.FieldProofs[0], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	// Validate roles
	valid, err = documents.ValidateProof(proof.FieldProofs[1], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := identity.NewDIDFromBytes(proof.FieldProofs[1].Value)
	assert.NoError(t, err)
	assert.True(t, i.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(proof.FieldProofs[2], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	// validate line item
	valid, err = documents.ValidateProof(proof.FieldProofs[3], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	valid, err = documents.ValidateProof(proof.FieldProofs[4], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestInvoice_CreateNFTProofs(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	acc := tc.(*configstore.Account)
	acc.IdentityID = defaultDID[:]
	assert.NoError(t, err)

	i, _ := CreateInvoiceWithEmbedCD(t, nil, did, []identity.DID{defaultDID})
	tt := time.Now()
	i.Data.DateDue = &tt
	i.Data.Status = "unpaid"
	assert.NoError(t, err)
	sig, err := acc.SignMsg([]byte{0, 1, 2, 3})
	assert.NoError(t, err)
	i.AppendSignatures(sig)
	dataRoot := calculateBasicDataRoot(t, i)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)

	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(defaultDID[:], keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRootField := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, signerId)
	proofFields := []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due", "invoice.sender", "invoice.status", signingRootField, signatureSender, documents.CDTreePrefix + ".next_version"}
	proof, err := i.CreateProofs(proofFields)
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree := getDocumentRootTree(t, i)
	assert.Len(t, proofFields, 8)

	nodeHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate invoice_gross_amount
	valid, err := documents.ValidateProof(proof.FieldProofs[0], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	// Validate signing_root
	valid, err = tree.ValidateProof(proof.FieldProofs[5])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate signature
	signaturesTree, err := i.CoreDocument.GetSignaturesDataTree()
	assert.NoError(t, err)
	valid, err = signaturesTree.ValidateProof(proof.FieldProofs[6])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate next_version
	valid, err = documents.ValidateProof(proof.FieldProofs[7], dataRoot, nodeHash, sha3.NewKeccak256())
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestInvoiceModel_createProofsFieldDoesNotExist(t *testing.T) {
	i, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	_, err := i.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestInvoiceModel_GetDocumentID(t *testing.T) {
	i, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	assert.Equal(t, i.CoreDocument.ID(), i.ID())
}

func TestInvoiceModel_getDocumentDataTree(t *testing.T) {
	na := new(documents.Decimal)
	assert.NoError(t, na.SetString("2"))
	ga := new(documents.Decimal)
	assert.NoError(t, ga.SetString("2"))
	i, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
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

func TestInvoice_CollaboratorCanUpdate(t *testing.T) {
	inv, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	id1 := did
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
	data := oldInv.Data
	dec, err := documents.NewDecimal("55")
	assert.NoError(t, err)
	data.GrossAmount = dec
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	err = inv.unpackFromUpdatePayloadOld(inv, documents.UpdatePayload{
		DocumentID: inv.ID(),
		CreatePayload: documents.CreatePayload{
			Data: d,
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []identity.DID{id3},
			},
		},
	})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldInv.CollaboratorCanUpdate(inv, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldInv.CollaboratorCanUpdate(inv, id2))

	// update the id3 rules to update only total amount
	inv.CoreDocument.Document.TransitionRules[3].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT
	inv.CoreDocument.Document.TransitionRules[3].Field = append(compactPrefix(), 0, 0, 0, 18)
	assert.NoError(t, testRepo().Create(id1[:], inv.CurrentVersion(), inv))

	// fetch the document
	model, err = testRepo().Get(id1[:], inv.CurrentVersion())
	assert.NoError(t, err)
	oldInv = model.(*Invoice)
	data = oldInv.Data
	dec, err = documents.NewDecimal("55")
	assert.NoError(t, err)
	data.GrossAmount = dec
	data.Currency = "INR"
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	err = inv.unpackFromUpdatePayloadOld(inv, documents.UpdatePayload{
		DocumentID:    inv.ID(),
		CreatePayload: documents.CreatePayload{Data: d},
	})
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
	attr, err := documents.NewStringAttribute(label, documents.AttrString, value)
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
	attr, err := documents.NewStringAttribute(label, documents.AttrString, value)
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
	inv, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
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
	err := loadData(payload.Data, &inv.Data)
	assert.Error(t, err)
}

func calculateBasicDataRoot(t *testing.T, i *Invoice) []byte {
	dataLeaves, err := i.getDataLeaves()
	assert.NoError(t, err)
	trees, _, err := i.CoreDocument.SigningDataTrees(i.DocumentType(), dataLeaves)
	assert.NoError(t, err)
	return trees[0].RootHash()
}

func getDocumentRootTree(t *testing.T, i *Invoice) *proofs.DocumentTree {
	dataLeaves, err := i.getDataLeaves()
	assert.NoError(t, err)
	tree, err := i.CoreDocument.DocumentRootTree(i.DocumentType(), dataLeaves)
	assert.NoError(t, err)
	return tree
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
	err := loadData(payload.Data, &inv.Data)
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
	ctx := context.Background()

	// invalid data
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	payload.Data = invalidDecimalData(t)
	err := inv.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvoiceInvalidData, err))

	// invalid attributes
	attr, err := documents.NewStringAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttributeType("some type")
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	payload.Data = validData(t)
	err = inv.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = inv.DeriveFromCreatePayload(ctx, payload)
	assert.NoError(t, err)
}

func TestInvoice_unpackFromUpdatePayloadOld(t *testing.T) {
	payload := documents.UpdatePayload{}
	old, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)
	inv := new(Invoice)

	// invalid data
	payload.Data = invalidDecimalData(t)
	err := inv.unpackFromUpdatePayloadOld(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvoiceInvalidData, err))

	// invalid attributes
	attr, err := documents.NewStringAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttributeType("some type")
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	payload.Data = validData(t)
	err = inv.unpackFromUpdatePayloadOld(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = inv.unpackFromUpdatePayloadOld(old, payload)
	assert.NoError(t, err)
}

func TestInvoice_unpackFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	old, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)

	// invalid data
	ctx := context.Background()
	payload.Data = invalidDecimalData(t)
	inv, err := old.DeriveFromUpdatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvoiceInvalidData, err))

	// invalid attributes
	attr, err := documents.NewStringAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttributeType("some type")
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	payload.Data = validData(t)
	_, err = old.DeriveFromUpdatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	inv, err = old.DeriveFromUpdatePayload(ctx, payload)
	assert.NoError(t, err)
	// check if patch worked
	assert.NotEqual(t, inv.GetData(), old.Data)
	assert.Equal(t, inv.GetData().(Data).Recipient.String(), "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.Equal(t, old.Data.Recipient.String(), "0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.Len(t, inv.GetData().(Data).LineItems, 1)

	// new data
	assert.Len(t, old.Data.Attachments, 0)
	assert.Len(t, inv.GetData().(Data).Attachments, 1)
}

func TestInvoice_Patch(t *testing.T) {
	payload := documents.UpdatePayload{}
	inv, _ := CreateInvoiceWithEmbedCD(t, nil, did, nil)

	// invalid data
	payload.Data = invalidDecimalData(t)
	err := inv.Patch(payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvoiceInvalidData, err))

	// valid
	payload.Data = validData(t)
	attr, err := documents.NewStringAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = inv.Patch(payload)
	assert.NoError(t, err)
	assert.Equal(t, inv.Data.Recipient.String(), "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	collabs, err := inv.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, collabs.ReadWriteCollaborators, 0)
}
