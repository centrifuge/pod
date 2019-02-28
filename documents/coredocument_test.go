// +build unit

package documents

import (
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
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
	cd := newCoreDocument()

	// missing DocumentRoot
	c1 := testingidentity.GenerateRandomDID()
	c2 := testingidentity.GenerateRandomDID()
	c := []string{c1.String(), c2.String()}
	ncd, err := cd.PrepareNewVersion(c, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Document root is invalid")
	assert.Nil(t, ncd)

	//collaborators need to be hex string
	cd.Document.DocumentRoot = utils.RandomSlice(32)
	collabs := []string{"some ID"}
	ncd, err = cd.PrepareNewVersion(collabs, false)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(identity.ErrMalformedAddress, err))
	assert.Nil(t, ncd)

	// successful preparation of new version upon addition of DocumentRoot
	ncd, err = cd.PrepareNewVersion(c, false)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	cs, err := ncd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, c1)
	assert.Contains(t, cs, c2)
	assert.Nil(t, ncd.Document.CoredocumentSalts)

	ncd, err = cd.PrepareNewVersion(c, true)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	cs, err = ncd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, c1)
	assert.Contains(t, cs, c2)
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

	_, err = cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err)

	_, err = cd.CalculateDocumentRoot()
	assert.Nil(t, err)

	hashes, err := cd.getSigningRootProofHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.Document.SigningRoot, hashes, cd.Document.DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetSignaturesTree(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd := newCoreDocument()
	cd.Document.EmbeddedData = docAny
	cd.Document.DataRoot = utils.RandomSlice(32)
	sig := &coredocumentpb.Signature{
		SignerId: utils.RandomSlice(identity.DIDLength),
		PublicKey: utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(52),
		Signature: utils.RandomSlice(32),
	}
	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{sig}
	err := cd.setSalts()
	assert.NoError(t, err)

	signatureTree, err := cd.getSignatureDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, signatureTree)

	lengthIdx, lengthLeaf := signatureTree.GetLeafByProperty(SignaturesTreePrefix+".signatures.length")
	assert.Equal(t, 0, lengthIdx)
	assert.NotNil(t, lengthLeaf)
	assert.Equal(t, SignaturesTreePrefix+".signatures.length", lengthLeaf.Property.ReadableName())
	assert.Equal(t, append(compactProperties(SignaturesTreePrefix), []byte{0, 0, 0, 1}...), lengthLeaf.Property.CompactName())

	signerKey := hexutil.Encode(sig.SignatureId)
	_, signerLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signerLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey), signerLeaf.Property.ReadableName())
	assert.Equal(t, append(compactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, append(sig.SignatureId , []byte{0, 0, 0, 2}...)...)...), signerLeaf.Property.CompactName())
	assert.Equal(t, sig.SignerId, signerLeaf.Value)
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

	_, leaf := tree.GetLeafByProperty(SigningTreePrefix + ".data_root")
	assert.NotNil(t, leaf)

	_, leaf = tree.GetLeafByProperty(SigningTreePrefix + ".cd_root")
	assert.NotNil(t, leaf)
}

// TestGetDocumentRootTree tests that the documentroottree is properly calculated
func TestGetDocumentRootTree(t *testing.T) {
	cd := newCoreDocument()

	sig := &coredocumentpb.Signature{
		SignerId: utils.RandomSlice(identity.DIDLength),
		PublicKey: utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(52),
		Signature: utils.RandomSlice(32),
	}
	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{sig}

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

	// Get some signer signature
	signerKey := hexutil.Encode(sig.SignatureId)
	_, signerLeaf := tree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signerLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey), signerLeaf.Property.ReadableName())
	assert.Equal(t, append(compactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, append(sig.SignatureId , []byte{0, 0, 0, 2}...)...)...), signerLeaf.Property.CompactName())
	assert.Equal(t, sig.SignerId, signerLeaf.Value)
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
	assert.NoError(t, cd.setSalts())
	cd.Document.DataRoot = testTree.RootHash()
	_, err = cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	_, err = cd.CalculateDocumentRoot()
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
			CDTreePrefix + ".document_identifier",
			true,
			6,
		},
		{
			"sample_field2",
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
	cd.addNewRule(role, coredocumentpb.Action_ACTION_READ)

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
	cd.addNewRule(role, coredocumentpb.Action_ACTION_READ)

	cs, err = cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, id1)
	assert.Contains(t, cs, id2)

	cs, err = cd.GetCollaborators(id2)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Contains(t, cs, id1)
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
	cd.addNewRule(role, coredocumentpb.Action_ACTION_READ)

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
