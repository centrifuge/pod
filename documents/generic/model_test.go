// +build unit

package generic

import (
	"crypto/sha256"
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
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
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
var did = testingidentity.GenerateRandomDID()

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
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestGeneric_PackCoreDocument(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	cd, err := g.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
}

func createCDWithEmbeddedGeneric(t *testing.T) (documents.Model, coredocumentpb.CoreDocument) {
	g := new(Generic)
	var err error
	cd, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{did}}, nil)
	assert.NoError(t, err)
	g.CoreDocument = cd
	g.GetTestCoreDocWithReset()
	_, err = g.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = g.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)
	ccd, err := g.PackCoreDocument()
	assert.NoError(t, err)
	return g, ccd
}

func TestGeneric_UnpackCoreDocument(t *testing.T) {
	var err error

	// embed data missing
	err = new(Generic).UnpackCoreDocument(coredocumentpb.CoreDocument{})
	assert.Error(t, err)

	// embed data type is wrong
	err = new(Generic).UnpackCoreDocument(coredocumentpb.CoreDocument{EmbeddedData: &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
	}})

	// successful
	g, cd := createCDWithEmbeddedGeneric(t)
	err = g.UnpackCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, g.ID(), g.ID())
	assert.Equal(t, g.CurrentVersion(), g.CurrentVersion())
	assert.Equal(t, g.PreviousVersion(), g.PreviousVersion())
	assert.Empty(t, g.GetData())
}

func TestGeneric_JSON(t *testing.T) {
	g, cd := createCDWithEmbeddedGeneric(t)
	cd, err := g.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := g.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	g = new(Generic)
	err = g.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := g.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged invoice variables")
	assert.Equal(t, cd, ncd)
}

func TestGeneric_calculateDataRoot(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)

	dr, err := g.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
}

func TestGeneric_CreateProofs(t *testing.T) {
	gm, cd := createCDWithEmbeddedGeneric(t)
	g := gm.(*Generic)
	rk := cd.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := g.CreateProofs([]string{pf, documents.CDTreePrefix + ".document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	if err != nil {
		return
	}

	tree, err := g.DocumentRootTree()
	assert.NoError(t, err)

	h := sha256.New()
	dataLeaves, err := g.getDataLeaves()
	assert.NoError(t, err)

	// Validate roles
	err = g.CoreDocument.ValidateDataProof(pf, g.DocumentType(), tree.RootHash(), proof[0], dataLeaves, h)
	assert.Nil(t, err)

	// Validate []byte value
	acc, err := identity.NewDIDFromBytes(proof[0].Value)
	assert.NoError(t, err)
	assert.True(t, g.AccountCanRead(acc))

	// Validate document_type
	err = g.CoreDocument.ValidateDataProof(documents.CDTreePrefix+".document_type", g.DocumentType(), tree.RootHash(), proof[1], dataLeaves, h)
	assert.Nil(t, err)
}

func TestGeneric_CreateNFTProofs(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	acc := tc.(*configstore.Account)
	acc.IdentityID = did[:]
	assert.NoError(t, err)
	gm, _ := createCDWithEmbeddedGeneric(t)
	g := gm.(*Generic)
	sigs, err := acc.SignMsg([]byte{0, 1, 2, 3})
	assert.NoError(t, err)
	g.AppendSignatures(sigs...)
	_, err = g.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = g.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)

	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(did[:], keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.DataRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, signerId)
	proofFields := []string{signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"}
	proof, err := g.CreateProofs(proofFields)
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := g.DocumentRootTree()
	assert.NoError(t, err)
	assert.Len(t, proofFields, 3)

	h := sha256.New()
	dataLeaves, err := g.getDataLeaves()
	assert.NoError(t, err)

	// Validate signing_root
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate signature
	err = g.ValidateDataProof(proofFields[1], g.DocumentType(), tree.RootHash(), proof[1], nil, h)
	assert.Nil(t, err)

	// Validate next_version
	err = g.ValidateDataProof(proofFields[2], g.DocumentType(), tree.RootHash(), proof[2], dataLeaves, h)
	assert.Nil(t, err)
}

func TestGeneric_getDataTree(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	tree, err := g.(*Generic).getDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("generic.scheme")
	assert.NotNil(t, leaf)
	assert.Equal(t, "generic.scheme", leaf.Property.ReadableName())
	assert.Equal(t, []byte(scheme), leaf.Value)
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
	testRepoGlobal.Register(&Generic{})
	return testRepoGlobal
}

func TestGeneric_CollaboratorCanUpdate(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	id1 := did
	id2 := testingidentity.GenerateRandomDID()

	// wrong type
	err := g.CollaboratorCanUpdate(new(mockModel), id1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))
	assert.NoError(t, testRepo().Create(id1[:], g.CurrentVersion(), g))

	// update the document
	model, err := testRepo().Get(id1[:], g.CurrentVersion())
	assert.NoError(t, err)
	oldGeneric := model.(*Generic)
	err = g.(*Generic).PrepareNewVersion(g, documents.CollaboratorsAccess{}, oldGeneric.Attributes)
	assert.NoError(t, err)

	_, err = g.CalculateDataRoot()
	assert.NoError(t, err)

	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldGeneric.CollaboratorCanUpdate(g, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldGeneric.CollaboratorCanUpdate(g, id2))
	assert.NoError(t, testRepo().Create(id1[:], g.CurrentVersion(), g))

	// fetch the document
	model, err = testRepo().Get(id1[:], g.CurrentVersion())
	assert.NoError(t, err)
	oldGeneric = model.(*Generic)
	err = g.(*Generic).PrepareNewVersion(g, documents.CollaboratorsAccess{}, oldGeneric.Attributes)
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldGeneric.CollaboratorCanUpdate(g, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldGeneric.CollaboratorCanUpdate(g, id2))
}

func TestGeneric_AddAttributes(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// success
	err = g.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.NoError(t, err)
	assert.True(t, g.AttributeExists(attr.Key))
	gattr, err := g.GetAttribute(attr.Key)
	assert.NoError(t, err)
	assert.Equal(t, attr, gattr)

	// fail
	attr.Value.Type = documents.AttributeType("some attr")
	err = g.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))
}

func TestGeneric_DeleteAttribute(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	label := "some key"
	value := "some value"
	attr, err := documents.NewAttribute(label, documents.AttrString, value)
	assert.NoError(t, err)

	// failed
	err = g.DeleteAttribute(attr.Key, true)
	assert.Error(t, err)

	// success
	assert.NoError(t, g.AddAttributes(documents.CollaboratorsAccess{}, true, attr))
	assert.True(t, g.AttributeExists(attr.Key))
	assert.NoError(t, g.DeleteAttribute(attr.Key, true))
	assert.False(t, g.AttributeExists(attr.Key))
}

func TestGeneric_GetData(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	data := g.GetData()
	assert.Equal(t, g.(*Generic).Data, data)
}

func marshallData(t *testing.T, m map[string]interface{}) []byte {
	data, err := json.Marshal(m)
	assert.NoError(t, err)
	return data
}

func validData(t *testing.T) []byte {
	d := map[string]interface{}{}
	return marshallData(t, d)
}

func TestGeneric_loadData(t *testing.T) {
	g := new(Generic)
	payload := documents.CreatePayload{}

	// valid data
	payload.Data = validData(t)
	err := g.loadData(payload.Data)
	assert.NoError(t, err)
	data := g.GetData().(Data)
	assert.Empty(t, data)
}

func TestGeneric_unpackFromCreatePayload(t *testing.T) {
	payload := documents.CreatePayload{}
	g := new(Generic)

	// invalid data
	err := g.unpackFromCreatePayload(did, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected end of JSON input")

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
	err = g.unpackFromCreatePayload(did, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = g.unpackFromCreatePayload(did, payload)
	assert.NoError(t, err)
}

func TestGeneric_unpackFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	old, _ := createCDWithEmbeddedGeneric(t)
	g := new(Generic)

	// invalid data
	err := g.unpackFromUpdatePayload(old.(*Generic), payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected end of JSON input")

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
	err = g.unpackFromUpdatePayload(old.(*Generic), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = g.unpackFromUpdatePayload(old.(*Generic), payload)
	assert.NoError(t, err)
}
