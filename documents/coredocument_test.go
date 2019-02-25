// +build unit

package documents

import (
	"crypto/sha256"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
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
	ctx[identity.BootstrappedIDService] = &testingcommons.MockIDService{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	ConfigService = ctx[config.BootstrappedConfigStorage].(config.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
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

func TestCoreDocument_PrepareNewVersion(t *testing.T) {
	cd := newCoreDocument()

	//collaborators need to be hex string
	collabs := []string{"some ID"}
	ncd, err := cd.PrepareNewVersion(collabs, false)
	assert.Error(t, err)
	assert.Nil(t, ncd)

	// missing DocumentRoot
	c1 := utils.RandomSlice(6)
	c2 := utils.RandomSlice(6)
	c := []string{hexutil.Encode(c1), hexutil.Encode(c2)}
	ncd, err = cd.PrepareNewVersion(c, false)
	assert.Error(t, err)
	assert.Nil(t, ncd)

	// successful preparation of new version upon addition of DocumentRoot
	cd.Document.DocumentRoot = utils.RandomSlice(32)
	ncd, err = cd.PrepareNewVersion(c, false)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	assert.Len(t, ncd.Document.Collaborators, 2)
	assert.Nil(t, ncd.Document.CoredocumentSalts)

	ncd, err = cd.PrepareNewVersion(c, true)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	assert.Len(t, ncd.Document.Collaborators, 2)
	assert.NotNil(t, ncd.Document.CoredocumentSalts)

	assert.Equal(t, cd.Document.NextVersion, ncd.Document.CurrentVersion)
	assert.Equal(t, cd.Document.CurrentVersion, ncd.Document.PreviousVersion)
	assert.Equal(t, cd.Document.DocumentIdentifier, ncd.Document.DocumentIdentifier)
	assert.Equal(t, cd.Document.DocumentRoot, ncd.Document.PreviousRoot)
}

func TestGetSigningProofHashes(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd := newCoreDocument()
	cd.Document.EmbeddedData = docAny
	cd.Document.DataRoot = utils.RandomSlice(32)
	err := cd.setSalts()
	assert.NoError(t, err)

	_, err = cd.SigningRoot(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err)

	_, err = cd.DocumentRoot()
	assert.Nil(t, err)

	hashes, err := cd.getSigningRootProofHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.Document.SigningRoot, hashes, cd.Document.DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDocumentSigningTree(t *testing.T) {
	cd := newCoreDocument()

	// no data root
	_, err := cd.signingRootTree(documenttypes.InvoiceDataTypeUrl)
	assert.Error(t, err)

	// successful tree generation

	cd.Document.DataRoot = utils.RandomSlice(32)
	assert.NoError(t, cd.setSalts())
	tree, err := cd.signingRootTree(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err)
	assert.NotNil(t, tree)

	_, leaf := tree.GetLeafByProperty("data_root")
	assert.NotNil(t, leaf)

	_, leaf = tree.GetLeafByProperty("cd_root")
	assert.NotNil(t, leaf)
}

// TestGetDocumentRootTree tests that the documentroottree is properly calculated
func TestGetDocumentRootTree(t *testing.T) {
	cd := newCoreDocument()

	// no signing root generated
	_, err := cd.DocumentRootTree()
	assert.Error(t, err)

	// successful Document root generation
	cd.Document.SigningRoot = utils.RandomSlice(32)
	tree, err := cd.DocumentRootTree()
	assert.NoError(t, err)
	_, leaf := tree.GetLeafByProperty("signing_root")
	assert.NotNil(t, leaf)
	assert.Equal(t, cd.Document.SigningRoot, leaf.Hash)
}

func TestCoreDocument_GenerateProofs(t *testing.T) {
	h := sha256.New()
	testTree := NewDefaultTree(nil)
	props := []proofs.Property{NewLeafProperty("sample_field", []byte{0, 0, 0, 200}), NewLeafProperty("sample_field2", []byte{0, 0, 0, 202})}
	compactProps := [][]byte{props[0].Compact, props[1].Compact}
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

	cd := newCoreDocument()
	cd.Document.EmbeddedData = docAny
	cd.Document.Collaborators = [][]byte{utils.RandomSlice(32), utils.RandomSlice(32)}
	assert.NoError(t, cd.setSalts())
	cd.Document.DataRoot = testTree.RootHash()
	_, err = cd.SigningRoot(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	_, err = cd.DocumentRoot()
	assert.NoError(t, err)

	cdTree, err := cd.documentTree(documenttypes.InvoiceDataTypeUrl)
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
			p, err := cd.CreateProofs(documenttypes.InvoiceDataTypeUrl, testTree, []string{test.fieldName})
			assert.NoError(t, err)
			assert.Equal(t, test.proofLength, len(p[0].SortedHashes))
			var l *proofs.LeafNode
			if test.fromCoreDoc {
				_, l = cdTree.GetLeafByProperty(test.fieldName)
			} else {
				_, l = testTree.GetLeafByProperty(test.fieldName)
				assert.Contains(t, compactProps, l.Property.CompactName())
			}
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes, cd.Document.DocumentRoot, h)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}
}

func TestCoreDocument_setSalts(t *testing.T) {
	cd := newCoreDocument()
	assert.Nil(t, cd.Document.CoredocumentSalts)

	assert.NoError(t, cd.setSalts())
	salts := cd.Document.CoredocumentSalts
	assert.Nil(t, cd.setSalts())
	assert.Equal(t, salts, cd.Document.CoredocumentSalts)
}
