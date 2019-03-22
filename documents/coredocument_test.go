// +build unit

package documents

import (
	"crypto/sha256"
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
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions/txv1"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

var ctx map[string]interface{}
var ConfigService config.Service
var cfg config.Configuration

func TestMain(m *testing.M) {
	ctx = make(map[string]interface{})
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		txv1.Bootstrapper{},
		&queue.Bootstrapper{},
		&anchors.Bootstrapper{},
		&Bootstrapper{},
	}
	ctx[identity.BootstrappedDIDService] = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDFactory] = &testingcommons.MockIdentityFactory{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	ConfigService = ctx[config.BootstrappedConfigStorage].(config.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)

	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func Test_fetchUniqueCollaborators(t *testing.T) {
	o1 := testingidentity.GenerateRandomDID()
	o2 := testingidentity.GenerateRandomDID()
	n1 := testingidentity.GenerateRandomDID()
	tests := []struct {
		old    []identity.DID
		new    []identity.DID
		result []identity.DID
	}{
		// when old cs are nil
		{
			new: []identity.DID{n1},
		},

		{
			old:    []identity.DID{o1, o2},
			result: []identity.DID{o1, o2},
		},

		{
			old:    []identity.DID{o1},
			new:    []identity.DID{n1},
			result: []identity.DID{o1},
		},

		{
			old:    []identity.DID{o1, n1},
			new:    []identity.DID{n1},
			result: []identity.DID{o1},
		},

		{
			old:    []identity.DID{o1, n1},
			new:    []identity.DID{o2},
			result: []identity.DID{o1, n1},
		},
	}

	for _, c := range tests {
		uc := filterCollaborators(c.old, c.new...)
		assert.Equal(t, c.result, uc)
	}
}

func TestCoreDocument_PrepareNewVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	h := sha256.New()
	h.Write(cd.GetTestCoreDocWithReset().CurrentPreimage)
	var expectedCurrentVersion []byte
	expectedCurrentVersion = h.Sum(expectedCurrentVersion)
	assert.Equal(t, expectedCurrentVersion, cd.GetTestCoreDocWithReset().CurrentVersion)

	// missing DocumentRoot
	c1 := testingidentity.GenerateRandomDID()
	c2 := testingidentity.GenerateRandomDID()
	c := []string{c1.String(), c2.String()}
	ncd, err := cd.PrepareNewVersion(c, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document root is invalid")
	assert.Nil(t, ncd)

	//collaborators need to be hex string
	cd.GetTestCoreDocWithReset().DocumentRoot = utils.RandomSlice(32)
	collabs := []string{"some ID"}
	ncd, err = cd.PrepareNewVersion(collabs, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(identity.ErrMalformedAddress, err))
	assert.Nil(t, ncd)

	// successful preparation of new version upon addition of DocumentRoot
	ncd, err = cd.PrepareNewVersion(c, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	cs, err := ncd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, c1)
	assert.Contains(t, cs, c2)
	h = sha256.New()
	h.Write(ncd.GetTestCoreDocWithReset().NextPreimage)
	var expectedNextVersion []byte
	expectedNextVersion = h.Sum(expectedNextVersion)
	assert.Equal(t, expectedNextVersion, ncd.GetTestCoreDocWithReset().NextVersion)

	ncd, err = cd.PrepareNewVersion(c, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	cs, err = ncd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, c1)
	assert.Contains(t, cs, c2)

	assert.Equal(t, cd.GetTestCoreDocWithReset().NextVersion, ncd.GetTestCoreDocWithReset().CurrentVersion)
	assert.Equal(t, cd.GetTestCoreDocWithReset().CurrentVersion, ncd.GetTestCoreDocWithReset().PreviousVersion)
	assert.Equal(t, cd.GetTestCoreDocWithReset().DocumentIdentifier, ncd.GetTestCoreDocWithReset().DocumentIdentifier)
	assert.Equal(t, cd.GetTestCoreDocWithReset().DocumentRoot, ncd.GetTestCoreDocWithReset().PreviousRoot)
	assert.Len(t, cd.GetTestCoreDocWithReset().Roles, 0)
	assert.Len(t, cd.GetTestCoreDocWithReset().ReadRules, 0)
	assert.Len(t, cd.GetTestCoreDocWithReset().TransitionRules, 0)
	assert.Len(t, ncd.GetTestCoreDocWithReset().Roles, 2)
	assert.Len(t, ncd.GetTestCoreDocWithReset().ReadRules, 1)
	assert.Len(t, ncd.GetTestCoreDocWithReset().TransitionRules, 2)
	assert.Len(t, ncd.GetTestCoreDocWithReset().Roles[0].Collaborators, 2)
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[0].Collaborators[0], c1[:])
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[0].Collaborators[1], c2[:])
	assert.Len(t, ncd.GetTestCoreDocWithReset().Roles[1].Collaborators, 2)
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[1].Collaborators[0], c1[:])
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[1].Collaborators[1], c2[:])
}

func TestGetSigningProofHash(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	cd.GetTestCoreDocWithReset().DataRoot = utils.RandomSlice(32)
	_, err = cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err)

	cd.GetTestCoreDocWithReset()
	_, err = cd.CalculateDocumentRoot()
	assert.Nil(t, err)

	signatureTree, err := cd.getSignatureDataTree()
	assert.Nil(t, err)

	valid, err := proofs.ValidateProofSortedHashes(cd.GetTestCoreDocWithReset().SigningRoot, [][]byte{signatureTree.RootHash()}, cd.GetTestCoreDocWithReset().DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetSignaturesTree(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	cd.GetTestCoreDocWithReset().DataRoot = utils.RandomSlice(32)
	sig := &coredocumentpb.Signature{
		SignerId:    utils.RandomSlice(identity.DIDLength),
		PublicKey:   utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(52),
		Signature:   utils.RandomSlice(32),
	}
	cd.GetTestCoreDocWithReset().SignatureData.Signatures = []*coredocumentpb.Signature{sig}
	signatureTree, err := cd.getSignatureDataTree()

	signatureRoot, err := cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, signatureTree)
	assert.Equal(t, signatureTree.RootHash(), signatureRoot)

	lengthIdx, lengthLeaf := signatureTree.GetLeafByProperty(SignaturesTreePrefix + ".signatures.length")
	assert.Equal(t, 0, lengthIdx)
	assert.NotNil(t, lengthLeaf)
	assert.Equal(t, SignaturesTreePrefix+".signatures.length", lengthLeaf.Property.ReadableName())
	assert.Equal(t, append(compactProperties(SignaturesTreePrefix), []byte{0, 0, 0, 1}...), lengthLeaf.Property.CompactName())

	signerKey := hexutil.Encode(sig.SignatureId)
	_, signerLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signerLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey), signerLeaf.Property.ReadableName())
	assert.Equal(t, append(compactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, append(sig.SignatureId, []byte{0, 0, 0, 2}...)...)...), signerLeaf.Property.CompactName())
	assert.Equal(t, sig.SignerId, signerLeaf.Value)
}

func TestGetDocumentSigningTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// no data root
	_, err = cd.signingRootTree(documenttypes.InvoiceDataTypeUrl)
	assert.Error(t, err)

	// successful tree generation
	cd.GetTestCoreDocWithReset().DataRoot = utils.RandomSlice(32)
	tree, err := cd.signingRootTree(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err)
	assert.NotNil(t, tree)

	_, leaf := tree.GetLeafByProperty(SigningTreePrefix + ".data_root")
	for _, l := range tree.GetLeaves() {
		fmt.Printf("P: %s V: %v", l.Property.ReadableName(), l.Value)
	}
	assert.NotNil(t, leaf)

	_, leaf = tree.GetLeafByProperty(SigningTreePrefix + ".cd_root")
	assert.NotNil(t, leaf)
}

// TestGetDocumentRootTree tests that the documentroottree is properly calculated
func TestGetDocumentRootTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignerId:    utils.RandomSlice(identity.DIDLength),
		PublicKey:   utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(52),
		Signature:   utils.RandomSlice(32),
	}
	cd.GetTestCoreDocWithReset().SignatureData.Signatures = []*coredocumentpb.Signature{sig}

	// no signing root generated
	_, err = cd.DocumentRootTree()
	assert.Error(t, err)

	// successful document root generation
	cd.GetTestCoreDocWithReset().SigningRoot = utils.RandomSlice(32)
	tree, err := cd.DocumentRootTree()
	assert.NoError(t, err)
	_, leaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SigningRootField))
	assert.NotNil(t, leaf)
	assert.Equal(t, cd.GetTestCoreDocWithReset().SigningRoot, leaf.Hash)

	// Get signaturesLeaf
	_, signaturesLeaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField))
	assert.NotNil(t, signaturesLeaf)
	assert.Equal(t, fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField), signaturesLeaf.Property.ReadableName())
	assert.Equal(t, append(compactProperties(DRTreePrefix), compactProperties(SignaturesRootField)...), signaturesLeaf.Property.CompactName())
}

func TestCoreDocument_GenerateProofs(t *testing.T) {
	h := sha256.New()
	cd, err := newCoreDocument()
	testTree := cd.DefaultTreeWithPrefix("prefix", []byte{1, 0, 0, 0})
	props := []proofs.Property{NewLeafProperty("prefix.sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("prefix.sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
	compactProps := [][]byte{props[0].Compact, props[1].Compact}
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)
	err = testTree.Generate()
	assert.NoError(t, err)
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err = newCoreDocument()
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	cd.GetTestCoreDocWithReset().DataRoot = testTree.RootHash()
	_, err = cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	_, err = cd.CalculateDocumentRoot()
	assert.NoError(t, err)

	cdTree, err := cd.coredocTree(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	tests := []struct {
		fieldName   string
		fromCoreDoc bool
		proofLength int
	}{
		{
			"prefix.sample_field",
			false,
			3,
		},
		{
			CDTreePrefix + ".document_identifier",
			true,
			6,
		},
		{
			"prefix.sample_field2",
			false,
			3,
		},
		{
			CDTreePrefix + ".next_version",
			true,
			6,
		},
	}
	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			p, err := cd.CreateProofs(documenttypes.InvoiceDataTypeUrl, testTree, []string{test.fieldName})
			assert.NoError(t, err)
			assert.Equal(t, test.proofLength, len(p[0].SortedHashes))
			var l *proofs.LeafNode
			if test.fromCoreDoc {
				_, l = cdTree.GetLeafByProperty(test.fieldName)
				valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes[:4], cdTree.RootHash(), h)
				assert.NoError(t, err)
				assert.True(t, valid)
			} else {
				_, l = testTree.GetLeafByProperty(test.fieldName)
				assert.Contains(t, compactProps, l.Property.CompactName())
				valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes[:1], testTree.RootHash(), h)
				assert.NoError(t, err)
				assert.True(t, valid)
			}
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes, cd.GetTestCoreDocWithReset().DocumentRoot, h)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}
}

func TestCoreDocument_getCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	ids := []string{id1.String()}
	cd, err := NewCoreDocumentWithCollaborators(ids)
	assert.NoError(t, err)
	cs, err := cd.getCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id1)

	cs, err = cd.getCollaborators(coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, cs, 0)
	role := newRole()
	role.Collaborators = append(role.Collaborators, id2[:])
	cd.GetTestCoreDocWithReset().Roles = append(cd.GetTestCoreDocWithReset().Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)

	cs, err = cd.getCollaborators(coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id2)

	cs, err = cd.getCollaborators(coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, id1)
	assert.Contains(t, cs, id2)
}

func TestCoreDocument_GetCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()
	ids := []string{id1.String()}
	cd, err := NewCoreDocumentWithCollaborators(ids)
	assert.NoError(t, err)
	cs, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id1)

	cs, err = cd.GetCollaborators(id1)
	assert.NoError(t, err)
	assert.Len(t, cs, 0)

	role := newRole()
	role.Collaborators = append(role.Collaborators, id2[:])
	cd.GetTestCoreDocWithReset().Roles = append(cd.GetTestCoreDocWithReset().Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)

	cs, err = cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, id1)
	assert.Contains(t, cs, id2)

	cs, err = cd.GetCollaborators(id2)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Contains(t, cs, id1)

	role2 := newRole()
	role2.Collaborators = append(role.Collaborators, id3[:])
	cd.GetTestCoreDocWithReset().Roles = append(cd.GetTestCoreDocWithReset().Roles, role2)
	cd.addNewTransitionRule(role2.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
}

func TestCoreDocument_GetSignCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	ids := []string{id1.String()}
	cd, err := NewCoreDocumentWithCollaborators(ids)
	assert.NoError(t, err)
	cs, err := cd.GetSignerCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id1)

	cs, err = cd.GetSignerCollaborators(id1)
	assert.NoError(t, err)
	assert.Len(t, cs, 0)

	role := newRole()
	role.Collaborators = append(role.Collaborators, id2[:])
	cd.GetTestCoreDocWithReset().Roles = append(cd.GetTestCoreDocWithReset().Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)

	cs, err = cd.GetSignerCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Contains(t, cs, id1)
	assert.NotContains(t, cs, id2)

	cs, err = cd.GetSignerCollaborators(id2)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Contains(t, cs, id1)
}
