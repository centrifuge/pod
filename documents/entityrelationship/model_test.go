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
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/p2p"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
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
var (
	did = testingidentity.GenerateRandomDID()
)

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	txMan := &testingtx.MockTxManager{}
	ctx[transactions.BootstrappedService] = txMan
	done := make(chan bool)
	txMan.On("ExecuteWithinTX", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(transactions.NilTxID(), done, nil)
	ctx[nft.BootstrappedInvoiceUnpaid] = new(testingdocuments.MockRegistry)
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

func TestEntityRelationship_PackCoreDocument(t *testing.T) {
	er := new(EntityRelationship)
	assert.NoError(t, er.InitEntityRelationshipInput(testingdocuments.CreateEntityRelationshipPayload()))

	cd, err := er.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestEntityRelationship_JSON(t *testing.T) {
	er := new(EntityRelationship)
	assert.NoError(t, er.InitEntityRelationshipInput(testingdocuments.CreateEntityRelationshipPayload()))

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
	assert.Equal(t, model.getClientData(), model.getClientData(), entityRelationship.(*EntityRelationship).getClientData())
	assert.Equal(t, model.ID(), entityRelationship.ID())
	assert.Equal(t, model.CurrentVersion(), entityRelationship.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), entityRelationship.PreviousVersion())
}

func TestEntityRelationship_getClientData(t *testing.T) {
	entityRelationshipData := testingdocuments.CreateEntityRelationshipData()
	er := new(EntityRelationship)
	err := er.loadFromP2PProtobuf(&entityRelationshipData)
	assert.NoError(t, err)

	data := er.getClientData()
	assert.NotNil(t, data, "entity data should not be nil")
	assert.Equal(t, data.OwnerIdentity, er.OwnerIdentity.String())
	assert.Equal(t, data.TargetIdentity, er.TargetIdentity.String())
}

func TestEntityRelationship_InitEntityInput(t *testing.T) {
	// successful init
	data := &cliententitypb.EntityRelationshipData{
		OwnerIdentity:  testingidentity.GenerateRandomDID().String(),
		TargetIdentity: testingidentity.GenerateRandomDID().String(),
	}
	e := new(EntityRelationship)
	err := e.InitEntityRelationshipInput(&cliententitypb.EntityRelationshipCreatePayload{Data: data})
	assert.NoError(t, err)

	// invalid did
	e = new(EntityRelationship)
	data.TargetIdentity = "some random string"
	err = e.InitEntityRelationshipInput(&cliententitypb.EntityRelationshipCreatePayload{Data: data})
	assert.Contains(t, err.Error(), "malformed address provided")
}

func TestEntityRelationship_calculateDataRoot(t *testing.T) {
	m := new(EntityRelationship)
	err := m.InitEntityRelationshipInput(testingdocuments.CreateEntityRelationshipPayload())
	assert.NoError(t, err)
	m.GetTestCoreDocWithReset()

	dr, err := m.CalculateDataRoot()
	assert.NoError(t, err)
	assert.False(t, utils.IsEmptyByteSlice(dr))
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
	e := new(EntityRelationship)
	err := e.InitEntityRelationshipInput(testingdocuments.CreateEntityRelationshipPayload())
	assert.NoError(t, err)
	e.GetTestCoreDocWithReset()
	_, err = e.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateSigningRoot()
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
	assert.Equal(t, documenttypes.EntityRelationshipDocumentTypeUrl, e.DocumentType())
}

func TestEntityRelationship_getDocumentDataTree(t *testing.T) {
	e := createEntityRelationship(t)
	tree, err := e.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("entity_relationship.owner_identity")
	assert.NotNil(t, leaf)
	assert.Equal(t, "entity_relationship.owner_identity", leaf.Property.ReadableName())
}

func TestEntityRelationship_CollaboratorCanUpdate(t *testing.T) {
	er := createEntityRelationship(t)
	id1, err := identity.NewDIDFromString("0xed03Fa80291fF5DDC284DE6b51E716B130b05e20")
	assert.NoError(t, err)
	id2 := testingidentity.GenerateRandomDID()

	// wrong type
	err = er.CollaboratorCanUpdate(new(mockModel), id2)
	assert.Error(t, err)

	// update doc
	assert.NoError(t, testRepo().Create(id1[:], er.CurrentVersion(), er))
	model, err := testRepo().Get(id1[:], er.CurrentVersion())
	assert.NoError(t, err)

	// attempted updater is not owner of the relationship
	oldRelationship := model.(*EntityRelationship)
	assert.NoError(t, err)
	err = er.CollaboratorCanUpdate(oldRelationship, id2)
	assert.Contains(t, err.Error(), "identity attempting to update the document does not own this entity relationship")

	// attempted updater is owner of the relationship
	err = er.CollaboratorCanUpdate(oldRelationship, id1)
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

var testRepoGlobal documents.Repository

func testRepo() documents.Repository {
	if testRepoGlobal == nil {
		ldb, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
		if err != nil {
			panic(err)
		}
		testRepoGlobal = documents.NewDBRepository(leveldb.NewLevelDBRepository(ldb))
		testRepoGlobal.Register(&EntityRelationship{})
	}
	return testRepoGlobal
}

func createCDWithEmbeddedEntityRelationship(t *testing.T) (documents.Model, coredocumentpb.CoreDocument) {
	e := new(EntityRelationship)
	err := e.InitEntityRelationshipInput(testingdocuments.CreateEntityRelationshipPayload())
	assert.NoError(t, err)
	e.GetTestCoreDocWithReset()
	_, err = e.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = e.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := e.PackCoreDocument()
	assert.NoError(t, err)
	return e, cd
}
