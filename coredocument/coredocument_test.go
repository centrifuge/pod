// +build unit

package coredocument

import (
	"crypto/sha256"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestGetSigningProofHashes(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	cd := New()
	cd.EmbeddedData = docAny
	cd.DataRoot = utils.RandomSlice(32)
	err := FillSalts(cd)
	assert.NoError(t, err)

	err = CalculateSigningRoot(cd, cd.DataRoot)
	assert.Nil(t, err)

	err = CalculateDocumentRoot(cd)
	assert.Nil(t, err)

	hashes, err := getSigningRootProofHashes(cd)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.SigningRoot, hashes, cd.DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDataProofHashes(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	cd := New()
	cd.EmbeddedData = docAny
	cd.DataRoot = utils.RandomSlice(32)
	err := FillSalts(cd)
	assert.NoError(t, err)

	err = CalculateSigningRoot(cd, cd.DataRoot)
	assert.Nil(t, err)

	err = CalculateDocumentRoot(cd)
	assert.Nil(t, err)

	hashes, err := getDataProofHashes(cd, cd.DataRoot)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.DataRoot, hashes, cd.DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDocumentSigningTree(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	cd := New()
	cd.EmbeddedData = docAny
	err := FillSalts(cd)
	assert.NoError(t, err)
	tree, err := GetDocumentSigningTree(cd, cd.DataRoot)
	assert.Nil(t, err)
	assert.NotNil(t, tree)

	_, leaf := tree.GetLeafByProperty("data_root")
	assert.NotNil(t, leaf)

	_, leaf = tree.GetLeafByProperty("cd_root")
	assert.NotNil(t, leaf)
}

func TestGetDocumentSigningTree_EmptyEmbeddedData(t *testing.T) {
	cd := New()
	err := FillSalts(cd)
	assert.NoError(t, err)
	tree, err := GetDocumentSigningTree(cd, cd.DataRoot)
	assert.NotNil(t, err)
	assert.Nil(t, tree)
}

// TestGetDocumentRootTree tests that the documentroottree is properly calculated
func TestGetDocumentRootTree(t *testing.T) {
	cd := &coredocumentpb.CoreDocument{SignatureData: new(coredocumentpb.SignatureData), SigningRoot: []byte{0x72, 0xee, 0xb8, 0x88, 0x92, 0xf7, 0x6, 0x19, 0x82, 0x76, 0xe9, 0xe7, 0xfe, 0xcc, 0x33, 0xa, 0x66, 0x78, 0xd4, 0xa6, 0x5f, 0xf6, 0xa, 0xca, 0x2b, 0xe4, 0x17, 0xa9, 0xf6, 0x15, 0x67, 0xa1}}
	tree, err := GetDocumentRootTree(cd)

	// Manually constructing the two node tree:
	signaturesLengthLeaf := sha256.Sum256(append([]byte("signatures.length0"), make([]byte, 32)...))
	expectedRootHash := sha256.Sum256(append(signaturesLengthLeaf[:], cd.SigningRoot...))
	assert.Nil(t, err)
	assert.Equal(t, expectedRootHash[:], tree.RootHash())
}

func TestPrepareNewVersion(t *testing.T) {
	var doc coredocumentpb.CoreDocument
	id := utils.RandomSlice(32)
	cv := id
	nv := utils.RandomSlice(32)
	dr := utils.RandomSlice(32)
	doc = coredocumentpb.CoreDocument{}

	// failed new with collaborators
	collabs := []string{"some ID"}
	newDoc, err := PrepareNewVersion(doc, collabs)
	assert.Error(t, err)
	assert.Nil(t, newDoc)

	// missing doc identifier
	collabs = []string{"0x010203040506"}
	newDoc, err = PrepareNewVersion(doc, collabs)
	assert.NotNil(t, err)
	assert.Nil(t, newDoc)

	//missing current version
	doc.DocumentIdentifier = id
	newDoc, err = PrepareNewVersion(doc, collabs)
	assert.NotNil(t, err)
	assert.Nil(t, newDoc)

	doc.CurrentVersion = cv
	newDoc, err = PrepareNewVersion(doc, collabs)
	assert.NotNil(t, err)
	assert.Nil(t, newDoc)

	doc.NextVersion = nv
	newDoc, err = PrepareNewVersion(doc, collabs)
	assert.NotNil(t, err)
	assert.Nil(t, newDoc)

	doc.CurrentVersion = cv
	doc.NextVersion = nv
	doc.DocumentRoot = dr

	newDoc, err = PrepareNewVersion(doc, collabs)
	assert.Nil(t, err)

	// original document hasn't changed
	assert.Equal(t, cv, doc.CurrentVersion)

	// new document has changed
	assert.Equal(t, id, newDoc.DocumentIdentifier)
	assert.Equal(t, cv, newDoc.PreviousVersion)
	assert.Equal(t, nv, newDoc.CurrentVersion)
	assert.Equal(t, dr, newDoc.PreviousRoot)
	assert.Nil(t, newDoc.DocumentRoot)
}

func TestNewWithCollaborators(t *testing.T) {
	// messed up collaborators
	c := []string{"some id"}
	cd, err := NewWithCollaborators(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode collaborator")
	assert.Nil(t, cd)

	// success
	c1 := utils.RandomSlice(6)
	c2 := utils.RandomSlice(6)
	c = []string{hexutil.Encode(c1), hexutil.Encode(c2)}
	cd, err = NewWithCollaborators(c)
	assert.Nil(t, err)
	assert.NotNil(t, cd)
	assert.NotNil(t, cd.DocumentIdentifier)
	assert.NotNil(t, cd.CurrentVersion)
	assert.NotNil(t, cd.NextVersion)
	assert.NotNil(t, cd.Collaborators)
	assert.NotNil(t, cd.CoredocumentSalts)
	assert.Equal(t, [][]byte{c1, c2}, cd.Collaborators)
}

func TestCreateProofs(t *testing.T) {
	h := sha256.New()
	testTree := NewDefaultTree(nil)
	props := []proofs.Property{NewLeafProperty("sample_field", []byte{0, 0, 0, 200}), NewLeafProperty("sample_field2", []byte{0, 0, 0, 202})}
	err := testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)
	err = testTree.Generate()
	assert.NoError(t, err)
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	cd := New()
	cd.EmbeddedData = docAny
	cd.DocumentIdentifier = utils.RandomSlice(32)
	cd.NextVersion = utils.RandomSlice(32)
	cd.Collaborators = [][]byte{utils.RandomSlice(32), utils.RandomSlice(32)}
	err = FillSalts(cd)
	assert.NoError(t, err)
	err = CalculateSigningRoot(cd, testTree.RootHash())
	assert.NoError(t, err)
	err = CalculateDocumentRoot(cd)
	assert.NoError(t, err)
	cdTree, err := GetCoreDocTree(cd)
	assert.NoError(t, err)
	tests := []struct {
		fieldName   string
		fromCoreDoc bool
		proofLength int
	}{
		{
			"sample_field",
			false,
			3,
		},
		{
			"document_identifier",
			true,
			6,
		},
		{
			"sample_field2",
			false,
			3,
		},
		{
			"collaborators[0]",
			true,
			6,
		},
	}
	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			p, err := CreateProofs(testTree, cd, []string{test.fieldName})
			assert.NoError(t, err)
			assert.Equal(t, test.proofLength, len(p[0].SortedHashes))
			var l *proofs.LeafNode
			if test.fromCoreDoc {
				_, l = cdTree.GetLeafByProperty(test.fieldName)
			} else {
				_, l = testTree.GetLeafByProperty(test.fieldName)
			}
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes, cd.DocumentRoot, h)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}

}

func TestGetExternalCollaborators(t *testing.T) {
	c1 := utils.RandomSlice(6)
	c2 := utils.RandomSlice(6)
	c := []string{hexutil.Encode(c1), hexutil.Encode(c2)}
	cd, err := NewWithCollaborators(c)
	assert.Equal(t, [][]byte{c1, c2}, cd.Collaborators)
	self, err := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)
	collaborators, err := GetExternalCollaborators(self.ID, cd)
	assert.Nil(t, err)
	assert.NotNil(t, collaborators)
	assert.Equal(t, [][]byte{c1, c2}, collaborators)
}

func TestGetExternalCollaborators_WrongIDFormat(t *testing.T) {
	c1 := utils.RandomSlice(6)
	c2 := utils.RandomSlice(6)
	c := []string{hexutil.Encode(c1), hexutil.Encode(c2)}
	cd, err := NewWithCollaborators(c)
	assert.Equal(t, [][]byte{c1, c2}, cd.Collaborators)
	cd.Collaborators[1] = utils.RandomSlice(5)
	self, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	_, err = GetExternalCollaborators(self.ID, cd)
	assert.NotNil(t, err)
}

func Test_fetchUniqueCollaborators(t *testing.T) {
	tests := []struct {
		old    [][]byte
		new    []string
		result []identity.CentID
		err    bool
	}{
		{
			new:    []string{"0x010203040506"},
			result: []identity.CentID{{1, 2, 3, 4, 5, 6}},
		},

		{
			old:    [][]byte{{1, 2, 3, 2, 3, 1}},
			new:    []string{"0x010203040506"},
			result: []identity.CentID{{1, 2, 3, 4, 5, 6}},
		},

		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
			new: []string{"0x010203040506"},
		},

		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
		},

		// new collaborator with wrong format
		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
			new: []string{"0x0102030405"},
			err: true,
		},
	}

	for _, c := range tests {
		uc, err := fetchUniqueCollaborators(c.old, c.new)
		if err != nil {
			if c.err {
				continue
			}

			t.Fatal(err)
		}

		assert.Equal(t, c.result, uc)
	}
}

func TestPrepareNewVersion_read_rules(t *testing.T) {
	cd, err := NewWithCollaborators([]string{"0x010203040506"})
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)
	assert.Equal(t, cd.Roles[0].Collaborators, [][]byte{{1, 2, 3, 4, 5, 6}})
	cd.DocumentRoot = utils.RandomSlice(32)

	// prepare with zero collaborators
	ncd, err := PrepareNewVersion(*cd, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	assert.Len(t, ncd.ReadRules, 1)
	assert.Len(t, ncd.Roles, 1)
	assert.Equal(t, ncd.Roles[0].Collaborators, [][]byte{{1, 2, 3, 4, 5, 6}})

	// prepare with no unique one
	ncd, err = PrepareNewVersion(*cd, []string{"0x010203040506"})
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	assert.Len(t, ncd.ReadRules, 1)
	assert.Len(t, ncd.Roles, 1)
	assert.Equal(t, ncd.Roles[0].Collaborators, [][]byte{{1, 2, 3, 4, 5, 6}})

	// prepare with unique collaborators
	ncd, err = PrepareNewVersion(*cd, []string{"0x010202030203", "0x020301020304"})
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	assert.Len(t, ncd.ReadRules, 2)
	assert.Len(t, ncd.Roles, 2)
	assert.Equal(t, ncd.Collaborators, [][]byte{{1, 2, 3, 4, 5, 6}, {1, 2, 2, 3, 2, 3}, {2, 3, 1, 2, 3, 4}})
	assert.Equal(t, ncd.Roles[0].Collaborators, [][]byte{{1, 2, 3, 4, 5, 6}})
	assert.Equal(t, ncd.Roles[1].Collaborators, [][]byte{{1, 2, 2, 3, 2, 3}, {2, 3, 1, 2, 3, 4}})
}
