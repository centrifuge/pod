// +build unit

package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

var (
	did       = testingidentity.GenerateRandomDID()
	dIDBytes  = did[:]
	accountID = did[:]
)

type mockAnchorSrv struct {
	mock.Mock
	anchors.Service
}

func (m *mockAnchorSrv) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocumentRoot, error) {
	args := m.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocumentRoot)
	return docRoot, args.Error(1)
}

func (m *mockAnchorSrv) GetAnchorData(anchorID anchors.AnchorID) (docRoot anchors.DocumentRoot, anchoredTime time.Time, err error) {
	args := m.Called(anchorID)
	docRoot, _ = args.Get(0).(anchors.DocumentRoot)
	anchoredTime, _ = args.Get(1).(time.Time)
	return docRoot, anchoredTime, args.Error(2)
}

func TestMain(m *testing.M) {
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient
	jobMan := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobMan
	done := make(chan error)
	jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), done, nil)
	ctx[bootstrap.BootstrappedNFTService] = new(testingdocuments.MockRegistry)
	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv2.Bootstrapper{},
		&queue.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", did.String())
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestEntity_PackCoreDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	entity, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestEntity_JSON(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	entity, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := entity.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	entity = new(Entity)
	err = entity.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := entity.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged entity variables")
	assert.Equal(t, cd, ncd)
}

func TestEntityModel_UnpackCoreDocument(t *testing.T) {
	var model = new(Entity)
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
	entity, cd := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	err = model.UnpackCoreDocument(cd)
	assert.NoError(t, err)

	d := model.Data
	d1 := entity.Data
	assert.Equal(t, d.Addresses[0], d1.Addresses[0])
	assert.Equal(t, model.ID(), entity.ID())
	assert.Equal(t, model.CurrentVersion(), entity.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), entity.PreviousVersion())
}

func TestEntity_CreateProofs(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	e, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	rk := e.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := e.CreateProofs([]string{"entity.legal_name", pf, documents.CDTreePrefix + ".document_type"})
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	dataRoot := calculateBasicDataRoot(t, e)
	assert.NoError(t, err)

	nodeHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate entity_number
	valid, err := documents.ValidateProof(proof.FieldProofs[0], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles
	valid, err = documents.ValidateProof(proof.FieldProofs[1], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := identity.NewDIDFromBytes(proof.FieldProofs[1].Value)
	assert.NoError(t, err)
	assert.True(t, e.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(proof.FieldProofs[2], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestEntityModel_createProofsFieldDoesNotExist(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	e, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	_, err := e.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestEntityModel_GetDocumentID(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	e, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	assert.Equal(t, e.CoreDocument.ID(), e.ID())
}

func TestEntityModel_getDocumentDataTree(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	e, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	tree, err := e.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("entity.legal_name")
	assert.NotNil(t, leaf)
	assert.Equal(t, "entity.legal_name", leaf.Property.ReadableName())
}

func TestEntity_CollaboratorCanUpdate(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	entity, _ := CreateEntityWithEmbedCD(t, ctx, did, nil)
	id1 := did
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()

	// wrong type
	err := entity.CollaboratorCanUpdate(new(mockModel), id1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))
	assert.NoError(t, testRepo().Create(id1[:], entity.CurrentVersion(), entity))

	// update the document
	model, err := testRepo().Get(id1[:], entity.CurrentVersion())
	assert.NoError(t, err)
	oldEntity := model.(*Entity)
	data := oldEntity.Data
	data.LegalName = "new legal name"
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	err = entity.unpackFromUpdatePayload(entity, documents.UpdatePayload{
		DocumentID: entity.ID(),
		CreatePayload: documents.CreatePayload{
			Data: d,
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []identity.DID{id3},
			},
		},
	})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldEntity.CollaboratorCanUpdate(entity, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldEntity.CollaboratorCanUpdate(entity, id2))

	// update the id3 rules to update only legal fields
	entity.CoreDocument.Document.TransitionRules[3].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT
	entity.CoreDocument.Document.TransitionRules[3].Field = append(compactPrefix(), 0, 0, 0, 2)
	assert.NoError(t, testRepo().Create(id1[:], entity.CurrentVersion(), entity))

	// fetch the document
	model, err = testRepo().Get(id1[:], entity.CurrentVersion())
	assert.NoError(t, err)
	oldEntity = model.(*Entity)
	data = oldEntity.Data
	data.LegalName = "second new legal name"
	data.Contacts = nil
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	err = entity.unpackFromUpdatePayload(entity, documents.UpdatePayload{
		DocumentID: entity.ID(),
		CreatePayload: documents.CreatePayload{
			Data: d,
		},
	})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldEntity.CollaboratorCanUpdate(entity, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldEntity.CollaboratorCanUpdate(entity, id2))

	// id3 should pass with just one error since changing contacts is not allowed
	err = oldEntity.CollaboratorCanUpdate(entity, id3)
	assert.Error(t, err)
	assert.Equal(t, 5, errors.Len(err)) //five contact fields have been changed
	assert.Contains(t, err.Error(), "entity.contacts")

}

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func (m *mockModel) ID() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

var testRepoGlobal documents.Repository

func testRepo() documents.Repository {
	if testRepoGlobal != nil {
		return testRepoGlobal
	}

	ldb, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
	if err != nil {
		panic(err)
	}
	testRepoGlobal = documents.NewDBRepository(leveldb.NewLevelDBRepository(ldb))
	testRepoGlobal.Register(&Entity{})
	return testRepoGlobal
}

func TestEntity_AddAttributes(t *testing.T) {
	e, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	label := "some key"
	value := "some value"
	attr, err := documents.NewStringAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// success
	err = e.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.NoError(t, err)
	assert.True(t, e.AttributeExists(attr.Key))
	gattr, err := e.GetAttribute(attr.Key)
	assert.NoError(t, err)
	assert.Equal(t, attr, gattr)

	// fail
	attr.Value.Type = documents.AttributeType("some attr")
	err = e.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))
}

func TestEntity_DeleteAttribute(t *testing.T) {
	e, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	label := "some key"
	value := "some value"
	attr, err := documents.NewStringAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// failed
	err = e.DeleteAttribute(attr.Key, true)
	assert.Error(t, err)

	// success
	assert.NoError(t, e.AddAttributes(documents.CollaboratorsAccess{}, true, attr))
	assert.True(t, e.AttributeExists(attr.Key))
	assert.NoError(t, e.DeleteAttribute(attr.Key, true))
	assert.False(t, e.AttributeExists(attr.Key))
}

func TestEntity_GetData(t *testing.T) {
	e, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	data := e.GetData()
	assert.Equal(t, e.Data, data)
}

func marshallData(t *testing.T, m map[string]interface{}) []byte {
	data, err := json.Marshal(m)
	assert.NoError(t, err)
	return data
}

func emptyDIDData(t *testing.T) []byte {
	d := map[string]interface{}{
		"identity": "",
	}

	return marshallData(t, d)
}

func invalidDIDData(t *testing.T) []byte {
	d := map[string]interface{}{
		"identity": "1acdew123asdefres",
	}

	return marshallData(t, d)
}

func emptyPaymentDetail(t *testing.T) []byte {
	d := map[string]interface{}{
		"identity": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"payment_details": []map[string]interface{}{
			{},
			{"predefined": true},
		},
	}

	return marshallData(t, d)
}

func multiPaymentDetail(t *testing.T) []byte {
	d := map[string]interface{}{
		"identity": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"payment_details": []map[string]interface{}{
			{
				"predefined": true,
				"bank_payment_method": map[string]interface{}{
					"identifier": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				},
				"crypto_payment_method": map[string]interface{}{
					"identifier": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				},
			},
		},
	}

	return marshallData(t, d)
}

func validData(t *testing.T) []byte {
	d := map[string]interface{}{
		"legal_name": "Hello, World!",
		"payment_details": []map[string]interface{}{
			{
				"predefined": true,
				"bank_payment_method": map[string]interface{}{
					"identifier": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				},
			},
		},
	}

	return marshallData(t, d)
}

func validDataWithIdentity(t *testing.T) []byte {
	d := map[string]interface{}{
		"legal_name": "Hello, World!",
		"identity":   "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"payment_details": []map[string]interface{}{
			{
				"predefined": true,
				"bank_payment_method": map[string]interface{}{
					"identifier": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				},
			},
		},
	}

	return marshallData(t, d)
}

func checkEntityPayloadDataError(t *testing.T, e *Entity, payload documents.CreatePayload) {
	var d Data
	err := loadData(payload.Data, &d)
	assert.Error(t, err)
	e.Data = d
}

func TestEntity_loadData(t *testing.T) {
	e := new(Entity)
	payload := documents.CreatePayload{}

	// empty did data
	payload.Data = emptyDIDData(t)
	checkEntityPayloadDataError(t, e, payload)

	// invalid did data
	payload.Data = invalidDIDData(t)
	checkEntityPayloadDataError(t, e, payload)

	// empty payment detail
	payload.Data = emptyPaymentDetail(t)
	checkEntityPayloadDataError(t, e, payload)

	// multiple payment detail
	payload.Data = multiPaymentDetail(t)
	checkEntityPayloadDataError(t, e, payload)

	// valid data
	payload.Data = validData(t)
	var d Data
	err := loadData(payload.Data, &d)
	assert.NoError(t, err)
	e.Data = d
	data := e.GetData().(Data)
	assert.Equal(t, data.LegalName, "Hello, World!")
	assert.Len(t, data.PaymentDetails, 1)
	assert.NotNil(t, data.PaymentDetails[0].BankPaymentMethod)
	assert.Nil(t, data.PaymentDetails[0].CryptoPaymentMethod)
	assert.Nil(t, data.PaymentDetails[0].OtherPaymentMethod)
	assert.True(t, data.PaymentDetails[0].Predefined)
	assert.Equal(t, data.PaymentDetails[0].BankPaymentMethod.Identifier.String(), "0xbaeb33a61f05e6f269f1c4b4cff91a901b54daf7")
}

func TestEntity_DeriveFromCreatePayload(t *testing.T) {
	payload := documents.CreatePayload{}
	e := new(Entity)
	ctx := context.Background()

	// invalid data
	payload.Data = invalidDIDData(t)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	err := e.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEntityInvalidData, err))

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
	err = e.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = e.DeriveFromCreatePayload(ctx, payload)
	assert.NoError(t, err)
}

func TestInvoice_unpackFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	old, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	e := new(Entity)

	// invalid data
	payload.Data = invalidDIDData(t)
	err := e.unpackFromUpdatePayload(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEntityInvalidData, err))

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
	err = e.unpackFromUpdatePayload(old, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = e.unpackFromUpdatePayload(old, payload)
	assert.NoError(t, err)
}

func TestEntity_Patch(t *testing.T) {
	payload := documents.UpdatePayload{}
	doc, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)

	// invalid data
	payload.Data = invalidDIDData(t)
	err := doc.Patch(payload)
	assert.Error(t, err)

	// coredoc patch failed
	doc.CoreDocument.Status = documents.Committed
	payload.Data = validDataWithIdentity(t)
	err = doc.Patch(payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotInAllowedState, err))

	// success
	doc.CoreDocument.Status = documents.Pending
	err = doc.Patch(payload)
	assert.NoError(t, err)
}

func TestEntity_DeriveFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	doc, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
	ctx := context.Background()

	// invalid data
	payload.Data = invalidDIDData(t)
	_, err := doc.DeriveFromUpdatePayload(ctx, payload)
	assert.Error(t, err)

	// coredoc failed
	payload.Data = validDataWithIdentity(t)
	attr, err := documents.NewStringAttribute("test", documents.AttrString, "value")
	assert.NoError(t, err)
	val := attr.Value
	val.Type = documents.AttributeType("some type")
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	_, err = doc.DeriveFromUpdatePayload(ctx, payload)
	assert.Error(t, err)

	// Success
	payload.Attributes = nil
	gdoc, err := doc.DeriveFromUpdatePayload(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, gdoc)
}

func calculateBasicDataRoot(t *testing.T, e *Entity) []byte {
	dataLeaves, err := e.getDataLeaves()
	assert.NoError(t, err)
	trees, _, err := e.CoreDocument.SigningDataTrees(e.DocumentType(), dataLeaves)
	assert.NoError(t, err)
	return trees[0].RootHash()
}
