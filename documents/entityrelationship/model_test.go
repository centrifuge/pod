// +build unit

package entityrelationship

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
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var (
	did      = testingidentity.GenerateRandomDID()
	entityID = hexutil.Encode(utils.RandomSlice(32))
)

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
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", did.String())
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func CreateRelationshipData(t *testing.T) *entitypb.RelationshipData {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	selfDID, err := contextutil.AccountDID(ctxh)
	assert.NoError(t, err)
	return &entitypb.RelationshipData{
		OwnerIdentity:    selfDID.String(),
		TargetIdentity:   "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb",
		EntityIdentifier: hexutil.Encode(utils.RandomSlice(32)),
	}
}

func TestEntityRelationship_PackCoreDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	er := new(EntityRelationship)
	assert.NoError(t, er.InitEntityRelationshipInput(ctxh, entityID, CreateRelationshipData(t)))

	cd, err := er.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestEntityRelationship_PrepareNewVersion(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	selfDID, err := contextutil.AccountDID(ctxh)
	assert.NoError(t, err)

	m, _ := createCDWithEmbeddedEntityRelationship(t)
	old := m.(*EntityRelationship)
	data := &entitypb.RelationshipData{
		OwnerIdentity:  selfDID.String(),
		TargetIdentity: "random string",
	}
	err = old.PrepareNewVersion(old, data, nil)
	assert.Error(t, err)

	err = old.PrepareNewVersion(old, CreateRelationshipData(t), nil)
	assert.NoError(t, err)
}

func TestEntityRelationship_JSON(t *testing.T) {
	er := new(EntityRelationship)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	assert.NoError(t, er.InitEntityRelationshipInput(ctxh, entityID, CreateRelationshipData(t)))

	cd, err := er.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := er.JSON()
	assert.NoError(t, err)
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	er = new(EntityRelationship)
	err = er.FromJSON(jsonBytes)
	assert.NoError(t, err)

	ncd, err := er.PackCoreDocument()
	assert.NoError(t, err)
	assert.Equal(t, cd, ncd)
}

func TestEntityRelationship_UnpackCoreDocument(t *testing.T) {
	var model = new(EntityRelationship)

	// embed data missing
	err := model.UnpackCoreDocument(coredocumentpb.CoreDocument{})
	assert.Error(t, err)

	// embed data type is wrong
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{EmbeddedData: new(any.Any)})
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// embed data is wrong
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{
		EmbeddedData: &any.Any{
			Value:   utils.RandomSlice(32),
			TypeUrl: documenttypes.EntityRelationshipDataTypeUrl,
		},
	})
	assert.Error(t, err)

	// successful
	entityRelationship, cd := createCDWithEmbeddedEntityRelationship(t)
	err = model.UnpackCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, model.getRelationshipData(), model.getRelationshipData(), entityRelationship.(*EntityRelationship).getRelationshipData())
	assert.Equal(t, model.ID(), entityRelationship.ID())
	assert.Equal(t, model.CurrentVersion(), entityRelationship.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), entityRelationship.PreviousVersion())
}

func TestEntityRelationship_getRelationshipData(t *testing.T) {
	entityRelationship := testingdocuments.CreateRelationship()
	er := new(EntityRelationship)
	err := er.loadFromP2PProtobuf(entityRelationship)
	assert.NoError(t, err)

	data := er.getRelationshipData()
	assert.NotNil(t, data, "entity relationship data should not be nil")
	assert.Equal(t, data.OwnerIdentity, er.OwnerIdentity.String())
	assert.Equal(t, data.TargetIdentity, er.TargetIdentity.String())
}

func TestEntityRelationship_InitEntityInput(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	selfDID, err := contextutil.AccountDID(ctxh)
	assert.NoError(t, err)
	// successful init
	data := &entitypb.RelationshipData{
		OwnerIdentity:    selfDID.String(),
		TargetIdentity:   testingidentity.GenerateRandomDID().String(),
		EntityIdentifier: hexutil.Encode(utils.RandomSlice(32)),
	}
	e := new(EntityRelationship)
	err = e.InitEntityRelationshipInput(ctxh, entityID, data)
	assert.NoError(t, err)

	// invalid did
	e = new(EntityRelationship)
	data.TargetIdentity = "some random string"
	err = e.InitEntityRelationshipInput(ctxh, entityID, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "malformed address provided")
}

func TestEntityRelationship_calculateDataRoot(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	m := new(EntityRelationship)
	err := m.InitEntityRelationshipInput(ctxh, entityID, CreateRelationshipData(t))
	assert.NoError(t, err)
	m.GetTestCoreDocWithReset()

	dr, err := m.CalculateDataRoot()
	assert.NoError(t, err)
	assert.False(t, utils.IsEmptyByteSlice(dr))
}

func TestEntityRelationship_AddNFT(t *testing.T) {
	m := new(EntityRelationship)
	err := m.AddNFT(true, common.Address{}, nil)
	assert.Error(t, err)
}

func TestEntityRelationship_CreateNFTProofs(t *testing.T) {
	m := new(EntityRelationship)
	_, err := m.CreateNFTProofs(did, common.Address{}, utils.RandomSlice(32), true, true)
	assert.Error(t, err)
}

func TestEntityRelationship_CreateProofs(t *testing.T) {
	e := createEntityRelationship(t)
	rk := e.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := e.CreateProofs([]string{"entity_relationship.owner_identity", pf, documents.CDTreePrefix + ".document_type"})
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	tree, err := e.DocumentRootTree()
	assert.NoError(t, err)

	// Validate entity_number
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
	assert.True(t, e.AccountCanRead(acc))

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func createEntityRelationship(t *testing.T) *EntityRelationship {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	e := new(EntityRelationship)
	err := e.InitEntityRelationshipInput(ctxh, entityID, CreateRelationshipData(t))
	assert.NoError(t, err)
	e.GetTestCoreDocWithReset()
	_, err = e.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentRoot()
	assert.NoError(t, err)
	return e
}

func TestEntityRelationship_createProofsFieldDoesNotExist(t *testing.T) {
	e := createEntityRelationship(t)
	_, err := e.CreateProofs([]string{"nonexisting"})
	assert.Error(t, err)
}

func TestEntityRelationship_GetDocumentID(t *testing.T) {
	e := createEntityRelationship(t)
	assert.Equal(t, e.CoreDocument.ID(), e.ID())
}

func TestEntityRelationship_GetDocumentType(t *testing.T) {
	e := createEntityRelationship(t)
	assert.Equal(t, documenttypes.EntityRelationshipDataTypeUrl, e.DocumentType())
}

func TestEntityRelationship_getDocumentDataTree(t *testing.T) {
	e := createEntityRelationship(t)
	tree, err := e.getDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("entity_relationship.owner_identity")
	assert.NotNil(t, leaf)
	assert.Equal(t, "entity_relationship.owner_identity", leaf.Property.ReadableName())
}

func TestEntityRelationship_CollaboratorCanUpdate(t *testing.T) {
	er := createEntityRelationship(t)
	id1, err := identity.NewDIDFromString("0xed03Fa80291fF5DDC284DE6b51E716B130b05e20")
	assert.NoError(t, err)

	// wrong type
	err = er.CollaboratorCanUpdate(new(mockModel), id1)
	assert.Error(t, err)

	// update doc
	assert.NoError(t, testEntityRepo().Create(id1[:], er.CurrentVersion(), er))
	model, err := testEntityRepo().Get(id1[:], er.CurrentVersion())
	assert.NoError(t, err)

	// attempted updater is not owner of the relationship
	oldRelationship := model
	assert.NoError(t, err)
	err = er.CollaboratorCanUpdate(oldRelationship, id1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity attempting to update the document does not own this entity relationship")

	// attempted updater is owner of the relationship
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	selfDID, err := contextutil.AccountDID(ctxh)
	assert.NoError(t, err)
	err = er.CollaboratorCanUpdate(oldRelationship, selfDID)
	assert.NoError(t, err)
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

var testRepoGlobal repository
var testDocRepoGlobal documents.Repository

func testEntityRepo() repository {
	if testRepoGlobal != nil {
		return testRepoGlobal
	}

	ldb, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
	if err != nil {
		panic(err)
	}
	db := leveldb.NewLevelDBRepository(ldb)
	if testDocRepoGlobal == nil {
		testDocRepoGlobal = documents.NewDBRepository(db)
	}
	testRepoGlobal = newDBRepository(db, testDocRepoGlobal)
	testRepoGlobal.Register(&EntityRelationship{})
	return testRepoGlobal
}

func createCDWithEmbeddedEntityRelationship(t *testing.T) (documents.Model, coredocumentpb.CoreDocument) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	e := new(EntityRelationship)
	err := e.InitEntityRelationshipInput(ctxh, entityID, CreateRelationshipData(t))
	assert.NoError(t, err)
	e.GetTestCoreDocWithReset()
	_, err = e.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := e.PackCoreDocument()
	assert.NoError(t, err)
	return e, cd
}

func TestEntityRelationship_AddAttributes(t *testing.T) {
	e, _ := createCDWithEmbeddedEntityRelationship(t)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
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

func TestEntityRelationship_DeleteAttribute(t *testing.T) {
	e, _ := createCDWithEmbeddedEntityRelationship(t)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
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
