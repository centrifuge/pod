// +build unit

package entityrelationship

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
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
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common"
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
	did = testingidentity.GenerateRandomDID()
)

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

func TestEntityRelationship_PackCoreDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	er := CreateRelationship(t, ctxh)

	cd, err := er.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func TestEntityRelationship_JSON(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	er, _ := CreateCDWithEmbeddedEntityRelationship(t, ctxh)

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
	entityRelationship, cd := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	err = model.UnpackCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, model.Data, entityRelationship.(*EntityRelationship).Data)
	assert.Equal(t, model.ID(), entityRelationship.ID())
	assert.Equal(t, model.CurrentVersion(), entityRelationship.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), entityRelationship.PreviousVersion())
}

func TestEntityRelationship_getRelationshipData(t *testing.T) {
	e, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	er := e.(*EntityRelationship)
	data := er.GetData().(Data)
	assert.NotNil(t, data, "entity relationship data should not be nil")
	assert.Equal(t, data.OwnerIdentity, er.Data.OwnerIdentity)
	assert.Equal(t, data.TargetIdentity, er.Data.TargetIdentity)
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
	er, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	e := er.(*EntityRelationship)
	rk := e.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := e.CreateProofs([]string{"entity_relationship.owner_identity", pf, documents.CDTreePrefix + ".document_type"})
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	dataRoot := calculateBasicDataRoot(t, e)

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

func TestEntityRelationship_createProofsFieldDoesNotExist(t *testing.T) {
	e, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	_, err := e.CreateProofs([]string{"nonexisting"})
	assert.Error(t, err)
}

func TestEntityRelationship_GetDocumentID(t *testing.T) {
	er, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	e := er.(*EntityRelationship)
	assert.Equal(t, e.CoreDocument.ID(), e.ID())
}

func TestEntityRelationship_GetDocumentType(t *testing.T) {
	er, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	assert.Equal(t, documenttypes.EntityRelationshipDataTypeUrl, er.DocumentType())
}

func TestEntityRelationship_getDocumentDataTree(t *testing.T) {
	er, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	e := er.(*EntityRelationship)
	tree, err := e.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("entity_relationship.owner_identity")
	assert.NotNil(t, leaf)
	assert.Equal(t, "entity_relationship.owner_identity", leaf.Property.ReadableName())
}

func TestEntityRelationship_CollaboratorCanUpdate(t *testing.T) {
	er, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	id := testingidentity.GenerateRandomDID()

	// wrong type
	err := er.CollaboratorCanUpdate(new(mockModel), id)
	assert.Error(t, err)

	// update doc
	assert.NoError(t, testEntityRepo().Create(did[:], er.CurrentVersion(), er))

	// attempted updater is not owner of the relationship
	err = er.CollaboratorCanUpdate(er, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity attempting to update the document does not own this entity relationship")

	// attempted updater is owner of the relationship
	err = er.CollaboratorCanUpdate(er, *er.GetData().(Data).OwnerIdentity)
	assert.NoError(t, err)
}

type mockModel struct {
	documents.Document
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

func TestEntityRelationship_AddAttributes(t *testing.T) {
	e, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
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

func TestEntityRelationship_DeleteAttribute(t *testing.T) {
	e, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
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

func invalidData(t *testing.T) []byte {
	m := map[string]string{
		"target_identity": "",
	}

	d, err := json.Marshal(m)
	assert.NoError(t, err)
	return d
}

func validData(t *testing.T, self identity.DID) []byte {
	return validDataWithTargetDID(t, self, testingidentity.GenerateRandomDID())
}

func validDataWithTargetDID(t *testing.T, self, target identity.DID) []byte {
	m := map[string]string{
		"target_identity":   target.String(),
		"owner_identity":    self.String(),
		"entity_identifier": byteutils.HexBytes(utils.RandomSlice(32)).String(),
	}

	d, err := json.Marshal(m)
	assert.NoError(t, err)
	return d
}

func TestEntityRelationship_loadData(t *testing.T) {
	e := new(EntityRelationship)

	// invalid data
	d := invalidData(t)
	err := loadData(d, &e.Data)
	assert.Error(t, err)

	d = validData(t, testingidentity.GenerateRandomDID())
	err = loadData(d, &e.Data)
	assert.NoError(t, err)
}

func TestEntityRelationship_DeriveFromCreatePayload(t *testing.T) {
	e := new(EntityRelationship)
	var payload documents.CreatePayload
	ctx := context.Background()

	// invalid data
	payload.Data = invalidData(t)
	err := e.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)

	// missing account context
	payload.Data = validData(t, did)
	err = e.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), documents.ErrDocumentConfigAccountID.Error())

	// success
	ctx = testingconfig.CreateAccountContext(t, cfg)
	err = e.DeriveFromCreatePayload(ctx, payload)
	assert.NoError(t, err)
}

func TestEntityRelationship_DeriveFromUpdatePayload(t *testing.T) {
	e := new(EntityRelationship)
	_, err := e.DeriveFromUpdatePayload(context.Background(), documents.UpdatePayload{})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEntityRelationshipUpdate, err))
}

func TestEntityRelationship_Patch(t *testing.T) {
	e := CreateRelationship(t, testingconfig.CreateAccountContext(t, cfg))

	// invalid data
	d := invalidData(t)
	payload := documents.UpdatePayload{CreatePayload: documents.CreatePayload{Data: d}}
	err := e.Patch(payload)
	assert.Error(t, err)

	// core doc patch failed
	e.CoreDocument.Status = documents.Committed
	self := did
	target := testingidentity.GenerateRandomDID()
	assert.NotEqual(t, e.Data.TargetIdentity, &target)
	payload.Data = validDataWithTargetDID(t, self, target)
	err = e.Patch(payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotInAllowedState, err))

	// success
	assert.NotEqual(t, e.Data.TargetIdentity, &target)
	e.CoreDocument.Status = documents.Pending
	err = e.Patch(payload)
	assert.NoError(t, err)
	assert.Equal(t, e.Data.TargetIdentity, &target)
	assert.Equal(t, e.Data.OwnerIdentity, &self)
}

func TestEntityRelationship_revokeRelationship(t *testing.T) {
	old, _ := CreateCDWithEmbeddedEntityRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	e := old.(*EntityRelationship)
	er := new(EntityRelationship)

	// failed to remove token
	id := testingidentity.GenerateRandomDID()
	docID := utils.RandomSlice(32)
	payload := documents.AccessTokenParams{
		Grantee:            id.String(),
		DocumentIdentifier: hexutil.Encode(docID),
	}
	err := er.revokeRelationship(e, id)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrAccessTokenNotFound, err))

	// success
	cd, err := e.AddAccessToken(testingconfig.CreateAccountContext(t, cfg), payload)
	e.CoreDocument = cd
	err = er.revokeRelationship(e, id)
	assert.NoError(t, err)
}

func calculateBasicDataRoot(t *testing.T, e *EntityRelationship) []byte {
	dataLeaves, err := e.getDataLeaves()
	assert.NoError(t, err)
	trees, _, err := e.CoreDocument.SigningDataTrees(e.DocumentType(), dataLeaves)
	assert.NoError(t, err)
	return trees[0].RootHash()
}
