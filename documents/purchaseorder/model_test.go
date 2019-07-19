// +build unit

package purchaseorder

import (
	"crypto/sha256"
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
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	jobManager := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobManager
	done := make(chan bool)
	jobManager.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), done, nil)
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

func TestPurchaseOrder_PackCoreDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	po := new(PurchaseOrder)
	assert.NoError(t, po.unpackFromCreatePayload(did, CreatePOPayload(t, nil)))
	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestPurchaseOrder_JSON(t *testing.T) {
	po := new(PurchaseOrder)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	assert.NoError(t, po.unpackFromCreatePayload(did, CreatePOPayload(t, nil)))

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
	po, cd := CreatePOWithEmbedCD(t, nil, did, nil)
	assert.NoError(t, model.UnpackCoreDocument(cd))
	data := model.GetData()
	data1 := po.GetData()
	assert.Equal(t, data, data1)
	assert.Equal(t, model.ID(), po.ID())
	assert.Equal(t, model.CurrentVersion(), po.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), po.PreviousVersion())
}

func TestPOModel_calculateDataRoot(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	po := new(PurchaseOrder)
	assert.Nil(t, po.unpackFromCreatePayload(did, CreatePOPayload(t, nil)), "Init must pass")
	dr, err := po.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
}

func TestPOModel_CreateProofs(t *testing.T) {
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	assert.NotNil(t, po)
	rk := po.CoreDocument.GetTestCoreDocWithReset().Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := po.CreateProofs([]string{"po.number", pf, documents.CDTreePrefix + ".document_type", "po.line_items[0].status"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)

	signingRoot, err := po.CalculateSigningRoot()
	assert.NoError(t, err)

	// Validate po_number
	valid, err := documents.ValidateProof(proof[0], signingRoot, sha256.New())
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles collaborators
	valid, err = documents.ValidateProof(proof[1], signingRoot, sha256.New())
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := identity.NewDIDFromBytes(proof[1].Value)
	assert.NoError(t, err)
	assert.True(t, po.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(proof[2], signingRoot, sha256.New())
	assert.Nil(t, err)
	assert.True(t, valid)

	// validate line items
	valid, err = documents.ValidateProof(proof[3], signingRoot, sha256.New())
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestPOModel_createProofsFieldDoesNotExist(t *testing.T) {
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	_, err := po.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestPOModel_getDocumentDataTree(t *testing.T) {
	na := new(documents.Decimal)
	assert.NoError(t, na.SetString("2"))
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	po.Data.Number = "123"
	po.Data.TotalAmount = na
	tree, err := po.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("po.number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "po.number", leaf.Property.ReadableName())
	assert.Equal(t, []byte(po.Data.Number), leaf.Value)
}

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func TestPurchaseOrder_CollaboratorCanUpdate(t *testing.T) {
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	id1 := did
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
	data := oldPO.Data
	dec, err := documents.NewDecimal("50")
	assert.NoError(t, err)
	data.TotalAmount = dec
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	err = po.unpackFromUpdatePayload(po, documents.UpdatePayload{
		DocumentID: po.ID(),
		CreatePayload: documents.CreatePayload{
			Data: d,
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []identity.DID{id3},
			},
		},
	})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldPO.CollaboratorCanUpdate(po, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldPO.CollaboratorCanUpdate(po, id2))

	// update the id3 rules to update only total amount
	po.CoreDocument.Document.TransitionRules[3].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT
	po.CoreDocument.Document.TransitionRules[3].Field = append(compactPrefix(), 0, 0, 0, 18)
	assert.NoError(t, testRepo().Create(id1[:], po.CurrentVersion(), po))

	// fetch the document
	model, err = testRepo().Get(id1[:], po.CurrentVersion())
	assert.NoError(t, err)
	oldPO = model.(*PurchaseOrder)
	data = oldPO.Data
	dec, err = documents.NewDecimal("55")
	assert.NoError(t, err)
	data.TotalAmount = dec
	data.Currency = "INR"
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	err = po.unpackFromUpdatePayload(po, documents.UpdatePayload{
		DocumentID:    po.ID(),
		CreatePayload: documents.CreatePayload{Data: d},
	})
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

func TestPurchaseOrder_AddAttributes(t *testing.T) {
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// success
	err = po.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.NoError(t, err)
	assert.True(t, po.AttributeExists(attr.Key))
	gattr, err := po.GetAttribute(attr.Key)
	assert.NoError(t, err)
	assert.Equal(t, attr, gattr)

	// fail
	attr.Value.Type = documents.AttributeType("some attr")
	err = po.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))
}

func TestPurchaseOrder_DeleteAttribute(t *testing.T) {
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// failed
	err = po.DeleteAttribute(attr.Key, true)
	assert.Error(t, err)

	// success
	assert.NoError(t, po.AddAttributes(documents.CollaboratorsAccess{}, true, attr))
	assert.True(t, po.AttributeExists(attr.Key))
	assert.NoError(t, po.DeleteAttribute(attr.Key, true))
	assert.False(t, po.AttributeExists(attr.Key))
}

func TestPurchaseOrder_GetData(t *testing.T) {
	po, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	data := po.GetData()
	assert.Equal(t, po.Data, data)
}

func marshallData(t *testing.T, m map[string]interface{}) []byte {
	data, err := json.Marshal(m)
	assert.NoError(t, err)
	return data
}

func emptyDecimalData(t *testing.T) []byte {
	d := map[string]interface{}{
		"total_amount": "",
	}

	return marshallData(t, d)
}

func invalidDecimalData(t *testing.T) []byte {
	d := map[string]interface{}{
		"total_amount": "10.10.",
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
		"date_sent": "",
	}

	return marshallData(t, d)
}

func invalidTimeData(t *testing.T) []byte {
	d := map[string]interface{}{
		"date_sent": "1920-12-10",
	}

	return marshallData(t, d)
}

func validData(t *testing.T) []byte {
	d := map[string]interface{}{
		"number":         "12345",
		"status":         "unpaid",
		"total_amount":   "12.345",
		"recipient":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_sent":      "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_confirmed": "2019-05-24T14:48:44Z",        // rfc3339
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
		"number":         "12345",
		"status":         "unpaid",
		"total_amount":   "12.345",
		"recipient":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_sent":      "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_confirmed": "2019-05-24T14:48:44Z",        // rfc3339
		"currency":       "EUR",
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

func checkPOPayloadDataError(t *testing.T, po *PurchaseOrder, payload documents.CreatePayload) {
	err := po.loadData(payload.Data)
	assert.Error(t, err)
}

func TestPurchaseOrder_loadData(t *testing.T) {
	po := new(PurchaseOrder)
	payload := documents.CreatePayload{}

	// empty decimal data
	payload.Data = emptyDecimalData(t)
	checkPOPayloadDataError(t, po, payload)

	// invalid decimal data
	payload.Data = invalidDecimalData(t)
	checkPOPayloadDataError(t, po, payload)

	// empty did data
	payload.Data = emptyDIDData(t)
	checkPOPayloadDataError(t, po, payload)

	// invalid did data
	payload.Data = invalidDIDData(t)
	checkPOPayloadDataError(t, po, payload)

	// empty time data
	payload.Data = emptyTimeData(t)
	checkPOPayloadDataError(t, po, payload)

	// invalid time data
	payload.Data = invalidTimeData(t)
	checkPOPayloadDataError(t, po, payload)

	// valid data
	payload.Data = validData(t)
	err := po.loadData(payload.Data)
	assert.NoError(t, err)
	data := po.GetData().(Data)
	assert.Equal(t, data.Number, "12345")
	assert.Equal(t, data.Status, "unpaid")
	assert.Equal(t, data.TotalAmount.String(), "12.345")
	assert.Equal(t, data.Recipient.String(), "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.Equal(t, data.DateSent.UTC().Format(time.RFC3339Nano), "2019-05-24T14:48:44.308854Z")
	assert.Equal(t, data.DateConfirmed.UTC().Format(time.RFC3339), "2019-05-24T14:48:44Z")
	assert.Len(t, data.Attachments, 1)
	assert.Equal(t, data.Attachments[0].Name, "test")
	assert.Equal(t, data.Attachments[0].FileType, "pdf")
	assert.Equal(t, data.Attachments[0].Size, 1000202)
	assert.Equal(t, hexutil.Encode(data.Attachments[0].Checksum), "0xbaeb33a61f05e6f269f1c4b4cff91a901b54daf3")
	assert.Equal(t, hexutil.Encode(data.Attachments[0].Data), "0xbaeb33a61f05e6f269f1c4b4cff91a901b54daf7")
}

func TestPurchaseOrder_unpackFromCreatePayload(t *testing.T) {
	payload := documents.CreatePayload{}
	po := new(PurchaseOrder)

	// invalid data
	payload.Data = invalidDecimalData(t)
	err := po.unpackFromCreatePayload(did, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrPOInvalidData, err))

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
	err = po.unpackFromCreatePayload(did, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = po.unpackFromCreatePayload(did, payload)
	assert.NoError(t, err)
}

func TestPurchaseOrder_unpackFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	old, _ := CreatePOWithEmbedCD(t, nil, did, nil)
	po := new(PurchaseOrder)

	// invalid data
	payload.Data = invalidDecimalData(t)
	err := po.unpackFromUpdatePayload(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrPOInvalidData, err))

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
	err = po.unpackFromUpdatePayload(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = po.unpackFromUpdatePayload(old, payload)
	assert.NoError(t, err)
}
