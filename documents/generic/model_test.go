// +build unit

package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
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
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var did = testingidentity.GenerateRandomDID()

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
	_, err = g.CalculateSigningRoot()
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

func TestGeneric_CreateProofs(t *testing.T) {
	g, cd := createCDWithEmbeddedGeneric(t)
	gg := g.(*Generic)
	rk := cd.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := g.CreateProofs([]string{pf, documents.CDTreePrefix + ".document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	if err != nil {
		return
	}

	dataRoot := calculateBasicDataRoot(t, gg)

	nodeHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate roles
	valid, err := documents.ValidateProof(proof.FieldProofs[0], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := identity.NewDIDFromBytes(proof.FieldProofs[0].Value)
	assert.NoError(t, err)
	assert.True(t, g.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(proof.FieldProofs[1], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestAttributeProof(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	acc := tc.(*configstore.Account)
	acc.IdentityID = did[:]
	assert.NoError(t, err)
	g, _ := createCDWithEmbeddedGeneric(t)
	gg := g.(*Generic)
	var attrs []documents.Attribute
	loanAmount := "loanAmount"
	loanAmountValue := "100"
	attr0, err := documents.NewStringAttribute(loanAmount, documents.AttrInt256, loanAmountValue)
	assert.NoError(t, err)
	attrs = append(attrs, attr0)
	asIsValue := "asIsValue"
	asIsValueValue := "1000"
	attr1, err := documents.NewStringAttribute(asIsValue, documents.AttrInt256, asIsValueValue)
	assert.NoError(t, err)
	attrs = append(attrs, attr1)
	afterRehabValue := "afterRehabValue"
	afterRehabValueValue := "2000"
	attr2, err := documents.NewStringAttribute(afterRehabValue, documents.AttrInt256, afterRehabValueValue)
	assert.NoError(t, err)
	attrs = append(attrs, attr2)

	err = g.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)

	sig, err := acc.SignMsg([]byte{0, 1, 2, 3})
	assert.NoError(t, err)
	g.AppendSignatures(sig)
	dataRoot := calculateBasicDataRoot(t, gg)

	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(did[:], keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signatureSender := fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, signerId)
	attributeLoanAmount := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr0.Key.String())
	attributeAsIsVal := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr1.Key.String())
	attributeAfterRehabVal := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr2.Key.String())
	proofFields := []string{attributeLoanAmount, attributeAsIsVal, attributeAfterRehabVal, signatureSender}
	proof, err := g.CreateProofs(proofFields)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Len(t, proofFields, 4)

	nodeHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate loanAmount
	valid, err := documents.ValidateProof(proof.FieldProofs[0], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	// Validate asIsValue
	valid, err = documents.ValidateProof(proof.FieldProofs[1], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

	// Validate afterRehabValue
	valid, err = documents.ValidateProof(proof.FieldProofs[2], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.NoError(t, err)
	assert.True(t, valid)

}

func TestGeneric_CreateNFTProofs(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	acc := tc.(*configstore.Account)
	acc.IdentityID = did[:]
	assert.NoError(t, err)
	g, _ := createCDWithEmbeddedGeneric(t)
	gg := g.(*Generic)
	sig, err := acc.SignMsg([]byte{0, 1, 2, 3})
	assert.NoError(t, err)
	g.AppendSignatures(sig)
	dataRoot := calculateBasicDataRoot(t, gg)
	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)

	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(did[:], keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRootField := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, signerId)
	proofFields := []string{signingRootField, signatureSender, documents.CDTreePrefix + ".next_version"}
	proof, err := g.CreateProofs(proofFields)
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree := getDocumentRootTree(t, gg)
	assert.NoError(t, err)
	assert.Len(t, proofFields, 3)

	// Validate signing_root
	valid, err := tree.ValidateProof(proof.FieldProofs[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate signature
	signaturesTree, err := gg.CoreDocument.GetSignaturesDataTree()
	assert.NoError(t, err)
	valid, err = signaturesTree.ValidateProof(proof.FieldProofs[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	nodeHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate next_version
	valid, err = documents.ValidateProof(proof.FieldProofs[2], dataRoot, nodeHash, sha3.NewLegacyKeccak256())
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestGeneric_getDocumentDataTree(t *testing.T) {
	g, _ := createCDWithEmbeddedGeneric(t)
	tree, err := g.(*Generic).getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("generic.scheme")
	assert.NotNil(t, leaf)
	assert.Equal(t, "generic.scheme", leaf.Property.ReadableName())
	assert.Equal(t, []byte(Scheme), leaf.Value)
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

	_, err = g.CalculateSigningRoot()
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
	attr, err := documents.NewStringAttribute(label, documents.AttrString, value)
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
	attr, err := documents.NewStringAttribute(label, documents.AttrString, value)
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

func calculateBasicDataRoot(t *testing.T, g *Generic) []byte {
	dataLeaves, err := g.getDataLeaves()
	assert.NoError(t, err)
	trees, _, err := g.CoreDocument.SigningDataTrees(g.DocumentType(), dataLeaves)
	assert.NoError(t, err)
	return trees[0].RootHash()
}

func getDocumentRootTree(t *testing.T, g *Generic) *proofs.DocumentTree {
	dataLeaves, err := g.getDataLeaves()
	assert.NoError(t, err)
	tree, err := g.CoreDocument.DocumentRootTree(g.DocumentType(), dataLeaves)
	assert.NoError(t, err)
	return tree
}

func TestGeneric_DeriveFromCreatePayload(t *testing.T) {
	payload := documents.CreatePayload{Collaborators: documents.CollaboratorsAccess{
		ReadWriteCollaborators: []identity.DID{did},
	}}
	g := new(Generic)
	ctx := context.Background()

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
	err = g.DeriveFromCreatePayload(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = g.DeriveFromCreatePayload(ctx, payload)
	assert.NoError(t, err)
}

func TestGeneric_unpackFromUpdatePayload(t *testing.T) {
	payload := documents.UpdatePayload{}
	old, _ := createCDWithEmbeddedGeneric(t)
	g := new(Generic)

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
	err = g.unpackFromUpdatePayloadOld(old.(*Generic), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))

	// valid
	val.Type = documents.AttrString
	attr.Value = val
	payload.Attributes = map[documents.AttrKey]documents.Attribute{
		attr.Key: attr,
	}
	err = g.unpackFromUpdatePayloadOld(old.(*Generic), payload)
	assert.NoError(t, err)
}
