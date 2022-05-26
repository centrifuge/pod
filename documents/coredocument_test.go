//go:build unit
// +build unit

package documents

import (
	"bytes"
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
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/precise-proofs/proofs"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/blake2b"
)

var ctx map[string]interface{}
var cfg config.Configuration
var did = testingidentity.GenerateRandomDID()

func TestMain(m *testing.M) {
	ctx = make(map[string]interface{})
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient

	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobs.Bootstrapper{},
		&configstore.Bootstrapper{},
		&anchors.Bootstrapper{},
		&Bootstrapper{},
	}
	ctx[identity.BootstrappedDIDService] = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDFactory] = &identity.MockFactory{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("identityId", did.String())
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

func TestCoreDocument_CurrentVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	assert.Equal(t, cd.CurrentVersion(), cd.Document.CurrentVersion)
}

func TestCoreDocument_PreviousVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	assert.Equal(t, cd.PreviousVersion(), cd.Document.PreviousVersion)
}

func TestCoreDocument_NextVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	assert.Equal(t, cd.NextVersion(), cd.Document.NextVersion)
}

func TestCoreDocument_CurrentVersionPreimage(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	assert.Equal(t, cd.CurrentVersionPreimage(), cd.Document.CurrentPreimage)
}

func TestCoreDocument_Author(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	did := testingidentity.GenerateRandomDID()
	cd.Document.Author = did[:]
	a, err := cd.Author()
	assert.NoError(t, err)

	aID, err := identity.NewDIDFromBytes(cd.Document.Author)
	assert.NoError(t, err)
	assert.Equal(t, a, aID)
}

func TestCoreDocument_ID(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	assert.Equal(t, cd.Document.DocumentIdentifier, cd.ID())
}

func TestNewCoreDocumentWithCollaborators(t *testing.T) {
	did1 := testingidentity.GenerateRandomDID()
	did2 := testingidentity.GenerateRandomDID()
	c := &CollaboratorsAccess{
		ReadCollaborators:      []identity.DID{did1},
		ReadWriteCollaborators: []identity.DID{did2},
	}
	cd, err := NewCoreDocument([]byte("inv"), *c, nil)
	assert.NoError(t, err)

	collabs, err := cd.GetCollaborators(identity.DID{})
	assert.NoError(t, err)
	assert.Equal(t, did1, collabs.ReadCollaborators[0])
	assert.Equal(t, did2, collabs.ReadWriteCollaborators[0])
}

func TestNewCoreDocumentWithAccessToken(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	id := hexutil.Encode(cd.Document.DocumentIdentifier)
	did1 := testingidentity.GenerateRandomDID()

	// wrong granteeID format
	at := AccessTokenParams{
		Grantee:            "random string",
		DocumentIdentifier: id,
	}
	ncd, err := NewCoreDocumentWithAccessToken(ctxh, CompactProperties("inv"), at)
	assert.Error(t, err)

	// wrong docID
	at = AccessTokenParams{
		Grantee:            did1.String(),
		DocumentIdentifier: "random string",
	}
	ncd, err = NewCoreDocumentWithAccessToken(ctxh, CompactProperties("inv"), at)
	assert.Error(t, err)

	// correct access token params
	at = AccessTokenParams{
		Grantee:            did1.String(),
		DocumentIdentifier: id,
	}
	ncd, err = NewCoreDocumentWithAccessToken(ctxh, CompactProperties("inv"), at)
	assert.NoError(t, err)

	token := ncd.Document.AccessTokens[0]
	assert.Equal(t, token.DocumentIdentifier, cd.Document.DocumentIdentifier)
	assert.Equal(t, token.Grantee, did1[:])
	assert.NotEqual(t, cd.Document.DocumentIdentifier, ncd.Document.DocumentIdentifier)
}

func TestCoreDocument_PrepareNewVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	h, err := blake2b.New256(nil)
	assert.NoError(t, err)
	h.Write(cd.GetTestCoreDocWithReset().CurrentPreimage)
	var expectedCurrentVersion []byte
	expectedCurrentVersion = h.Sum(expectedCurrentVersion)
	assert.Equal(t, expectedCurrentVersion, cd.GetTestCoreDocWithReset().CurrentVersion)
	c1 := testingidentity.GenerateRandomDID()
	c2 := testingidentity.GenerateRandomDID()
	c3 := testingidentity.GenerateRandomDID()
	c4 := testingidentity.GenerateRandomDID()

	// successful preparation of new version with new read collaborators
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{[]identity.DID{c1, c2}, nil}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	rc, err := ncd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	assert.Contains(t, rc, c1)
	assert.Contains(t, rc, c2)
	h, err = blake2b.New256(nil)
	assert.NoError(t, err)
	h.Write(ncd.GetTestCoreDocWithReset().NextPreimage)
	var expectedNextVersion []byte
	expectedNextVersion = h.Sum(expectedNextVersion)
	assert.Equal(t, expectedNextVersion, ncd.GetTestCoreDocWithReset().NextVersion)

	// successful preparation of new version with read and write collaborators
	assert.NoError(t, err)
	ncd, err = cd.PrepareNewVersion([]byte("inv"), CollaboratorsAccess{[]identity.DID{c1, c2}, []identity.DID{c3, c4}}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	rc, err = ncd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NoError(t, err)
	assert.Len(t, rc, 4)
	assert.Contains(t, rc, c1)
	assert.Contains(t, rc, c2)
	assert.Contains(t, rc, c3)
	assert.Contains(t, rc, c4)
	wc, err := ncd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	assert.NoError(t, err)
	assert.Len(t, wc, 2)
	assert.Contains(t, wc, c3)
	assert.Contains(t, wc, c4)
	assert.NotContains(t, wc, c1)
	assert.NotContains(t, wc, c2)

	assert.Equal(t, cd.GetTestCoreDocWithReset().NextVersion, ncd.GetTestCoreDocWithReset().CurrentVersion)
	assert.Equal(t, cd.GetTestCoreDocWithReset().CurrentVersion, ncd.GetTestCoreDocWithReset().PreviousVersion)
	assert.Equal(t, cd.GetTestCoreDocWithReset().DocumentIdentifier, ncd.GetTestCoreDocWithReset().DocumentIdentifier)
	assert.Len(t, cd.GetTestCoreDocWithReset().Roles, 0)
	assert.Len(t, cd.GetTestCoreDocWithReset().ReadRules, 0)
	assert.Len(t, cd.GetTestCoreDocWithReset().TransitionRules, 0)
	assert.Len(t, ncd.GetTestCoreDocWithReset().Roles, 2)
	assert.Len(t, ncd.GetTestCoreDocWithReset().ReadRules, 1)
	assert.Len(t, ncd.GetTestCoreDocWithReset().TransitionRules, 2)
	assert.Len(t, ncd.GetTestCoreDocWithReset().Roles[0].Collaborators, 4)
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[0].Collaborators[0], c1[:])
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[0].Collaborators[1], c2[:])
	assert.Len(t, ncd.GetTestCoreDocWithReset().Roles[1].Collaborators, 2)
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[1].Collaborators[0], c3[:])
	assert.Equal(t, ncd.GetTestCoreDocWithReset().Roles[1].Collaborators[1], c4[:])
}

func TestCoreDocument_Patch(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// not in allowed status error
	err = cd.SetStatus(Committed)
	assert.NoError(t, err)
	ncd, err := cd.Patch(nil, CollaboratorsAccess{}, nil)
	assert.Error(t, err)
	assert.Nil(t, ncd)

	cd, err = newCoreDocument()
	assert.NoError(t, err)
	h, err := blake2b.New256(nil)
	assert.NoError(t, err)
	h.Write(cd.GetTestCoreDocWithReset().CurrentPreimage)
	var expectedCurrentVersion []byte
	expectedCurrentVersion = h.Sum(expectedCurrentVersion)
	assert.Equal(t, expectedCurrentVersion, cd.GetTestCoreDocWithReset().CurrentVersion)
	c1 := testingidentity.GenerateRandomDID()
	c2 := testingidentity.GenerateRandomDID()
	attr, err := NewStringAttribute("test", AttrString, "value")
	assert.NoError(t, err)
	attrs := map[AttrKey]Attribute{
		attr.Key: attr,
	}

	ncd, err = cd.Patch(nil, CollaboratorsAccess{[]identity.DID{c1, c2}, nil}, attrs)
	assert.NoError(t, err)
	assert.NotNil(t, ncd)
	assert.Equal(t, cd.CurrentVersion(), ncd.CurrentVersion())
	assert.Equal(t, cd.NextVersion(), ncd.NextVersion())
	collabs, err := ncd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, collabs.ReadCollaborators, 2)
	assert.Equal(t, c1, collabs.ReadCollaborators[0])
	assert.Len(t, ncd.Attributes, 1)
	assert.Equal(t, ncd.Attributes[attr.Key].Value, attr.Value)

	// Override existing collaborators and attribute
	c3 := testingidentity.GenerateRandomDID()
	attr, err = NewStringAttribute("test1", AttrString, "value1")
	assert.NoError(t, err)
	attrs = map[AttrKey]Attribute{
		attr.Key: attr,
	}
	oncd, err := ncd.Patch(nil, CollaboratorsAccess{[]identity.DID{c3}, nil}, attrs)
	assert.NoError(t, err)
	assert.NotNil(t, oncd)
	assert.Equal(t, cd.CurrentVersion(), ncd.CurrentVersion())
	assert.Equal(t, cd.NextVersion(), ncd.NextVersion())
	collabs, err = oncd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, collabs.ReadCollaborators, 1)
	assert.Equal(t, c3, collabs.ReadCollaborators[0])
	assert.Len(t, oncd.Attributes, 1)
	assert.Equal(t, oncd.Attributes[attr.Key].Value, attr.Value)
}

func TestCoreDocument_newRoleWithCollaborators(t *testing.T) {
	did1 := testingidentity.GenerateRandomDID()
	did2 := testingidentity.GenerateRandomDID()

	role := newRoleWithCollaborators(did1, did2)
	assert.Len(t, role.Collaborators, 2)
	assert.Equal(t, role.Collaborators[0], did1[:])
	assert.Equal(t, role.Collaborators[1], did2[:])
}

func TestCoreDocument_AddUpdateLog(t *testing.T) {
	did1 := testingidentity.GenerateRandomDID()
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	err = cd.AddUpdateLog(did1)
	assert.NoError(t, err)
	assert.Equal(t, cd.Document.Author, did1[:])
	assert.True(t, cd.Modified)
}

func TestGetSigningProofHash(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
	assert.NoError(t, err)
	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.Nil(t, err)

	cd.GetTestCoreDocWithReset()
	docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.Nil(t, err)

	signatureTree, err := cd.GetSignaturesDataTree()
	assert.Nil(t, err)
	h, err := blake2b.New256(nil)
	assert.NoError(t, err)
	valid, err := proofs.ValidateProofHashes(signingRoot, []*proofspb.MerkleHash{{Right: signatureTree.RootHash()}}, docRoot, h)
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
	sig := &coredocumentpb.Signature{
		SignerId:    utils.RandomSlice(identity.DIDLength),
		PublicKey:   utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(52),
		Signature:   utils.RandomSlice(32),
	}
	cd.GetTestCoreDocWithReset().SignatureData.Signatures = []*coredocumentpb.Signature{sig}
	signatureTree, err := cd.GetSignaturesDataTree()

	signatureRoot, err := cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, signatureTree)
	assert.Equal(t, signatureTree.RootHash(), signatureRoot)

	lengthIdx, lengthLeaf := signatureTree.GetLeafByProperty(SignaturesTreePrefix + ".signatures.length")
	assert.Equal(t, 0, lengthIdx)
	assert.NotNil(t, lengthLeaf)
	assert.Equal(t, SignaturesTreePrefix+".signatures.length", lengthLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), []byte{0, 0, 0, 1}...), lengthLeaf.Property.CompactName())

	signerKey := hexutil.Encode(sig.SignatureId)
	_, signerLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signerLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey), signerLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, sig.SignatureId...)...), signerLeaf.Property.CompactName())
	assert.Equal(t, byteutils.AddZeroBytesSuffix(sig.Signature, 66), signerLeaf.Value)
}

// TestGetDocumentRootTree tests that the document root tree is properly calculated
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
	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	// successful document root generation
	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)
	tree, err := cd.DocumentRootTree(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)
	_, leaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SigningRootField))
	assert.NotNil(t, leaf)
	assert.Equal(t, signingRoot, leaf.Hash)

	// Get signaturesLeaf
	_, signaturesLeaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField))
	assert.NotNil(t, signaturesLeaf)
	assert.Equal(t, fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField), signaturesLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(DRTreePrefix), CompactProperties(SignaturesRootField)...), signaturesLeaf.Property.CompactName())
}

func TestCoreDocument_GenerateProofs(t *testing.T) {
	h, err := blake2b.New256(nil)
	assert.NoError(t, err)
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	testTree, err := cd.DefaultTreeWithPrefix("prefix", []byte{1, 0, 0, 0})
	assert.NoError(t, err)
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
	dataRoot := calculateBasicDataRoot(t, cd, documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
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
			5,
		},
		{
			CDTreePrefix + ".document_identifier",
			true,
			5,
		},
		{
			"prefix.sample_field2",
			false,
			5,
		},
		{
			CDTreePrefix + ".next_version",
			true,
			5,
		},
	}
	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			p, err := cd.CreateProofs(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves(), []string{test.fieldName})
			assert.NoError(t, err)
			assert.Equal(t, test.proofLength, len(p.FieldProofs[0].SortedHashes))
			_, l := testTree.GetLeafByProperty(test.fieldName)
			if !test.fromCoreDoc {
				assert.Contains(t, compactProps, l.Property.CompactName())
			} else {
				_, l = cdTree.GetLeafByProperty(test.fieldName)
			}
			assert.NotNil(t, l)
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p.FieldProofs[0].SortedHashes, dataRoot, h)
			assert.NoError(t, err)
			assert.True(t, valid)
			docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
			assert.NoError(t, err)
			signRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
			assert.NoError(t, err)

			// Validate document root for basic data tree
			calcDocRoot := proofs.HashTwoValues(signRoot, p.SignaturesRoot, h)
			assert.Equal(t, docRoot, calcDocRoot)
		})
	}
}

func TestGetDataTreePrefix(t *testing.T) {
	cds, err := newCoreDocument()
	assert.NoError(t, err)
	testTree, err := cds.DefaultTreeWithPrefix("prefix", []byte{1, 0, 0, 0})
	assert.NoError(t, err)
	props := []proofs.Property{NewLeafProperty("prefix.sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("prefix.sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
	//compactProps := [][]byte{props[0].Compact, props[1].Compact}
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)
	err = testTree.Generate()
	assert.NoError(t, err)

	// nil length leaves
	prfx, err := getDataTreePrefix(nil)
	assert.Error(t, err)

	// zero length leaves
	prfx, err = getDataTreePrefix([]proofs.LeafNode{})
	assert.Error(t, err)

	// success
	prfx, err = getDataTreePrefix(testTree.GetLeaves())
	assert.NoError(t, err)
	assert.Equal(t, "prefix", prfx)

	// non-prefixed tree error
	testTree, err = cds.DefaultTreeWithPrefix("", []byte{1, 0, 0, 0})
	assert.NoError(t, err)
	props = []proofs.Property{NewLeafProperty("sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
	//compactProps := [][]byte{props[0].Compact, props[1].Compact}
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)
	err = testTree.Generate()
	assert.NoError(t, err)

	prfx, err = getDataTreePrefix(testTree.GetLeaves())
	assert.Error(t, err)
}

func TestCoreDocument_getReadCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	cas := CollaboratorsAccess{
		ReadWriteCollaborators: []identity.DID{id1},
	}
	cd, err := NewCoreDocument(nil, cas, nil)
	assert.NoError(t, err)
	cs, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id1)

	cs, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, cs, 0)
	role := newRoleWithCollaborators(id2)
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)

	cs, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id2)

	cs, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, id1)
	assert.Contains(t, cs, id2)
}

func TestCoreDocument_getWriteCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
	cd, err := NewCoreDocument([]byte("inv"), cas, nil)
	assert.NoError(t, err)
	cs, err := cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)

	role := newRoleWithCollaborators(id2)
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)

	cs, err = cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	assert.NoError(t, err)
	assert.Len(t, cs, 2)
	assert.Equal(t, cs[1], id2)
}

func TestCoreDocument_GetCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()
	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
	cd, err := NewCoreDocument(nil, cas, nil)
	assert.NoError(t, err)
	cs, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs.ReadCollaborators, 0)
	assert.Len(t, cs.ReadWriteCollaborators, 1)
	assert.Equal(t, cs.ReadWriteCollaborators[0], id1)

	role := newRoleWithCollaborators(id2)
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)

	cs, err = cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs.ReadCollaborators, 1)
	assert.Contains(t, cs.ReadCollaborators, id2)
	assert.Len(t, cs.ReadWriteCollaborators, 1)
	assert.Contains(t, cs.ReadWriteCollaborators, id1)

	cs, err = cd.GetCollaborators(id2)
	assert.NoError(t, err)
	assert.Len(t, cs.ReadCollaborators, 0)
	assert.Len(t, cs.ReadWriteCollaborators, 1)
	assert.Contains(t, cs.ReadWriteCollaborators, id1)

	role2 := newRoleWithCollaborators(id3)
	cd.Document.Roles = append(cd.Document.Roles, role2)
	cd.addNewTransitionRule(role2.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	cs, err = cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs.ReadCollaborators, 1)
	assert.Contains(t, cs.ReadCollaborators, id2)
	assert.Len(t, cs.ReadWriteCollaborators, 2)
	assert.Contains(t, cs.ReadWriteCollaborators, id1)
	assert.Contains(t, cs.ReadWriteCollaborators, id3)
}

func TestCoreDocument_GetSignCollaborators(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
	cd, err := NewCoreDocument(nil, cas, nil)
	assert.NoError(t, err)
	cs, err := cd.GetSignerCollaborators()
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0], id1)

	cs, err = cd.GetSignerCollaborators(id1)
	assert.NoError(t, err)
	assert.Len(t, cs, 0)

	role := newRoleWithCollaborators(id2)
	cd.Document.Roles = append(cd.Document.Roles, role)
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

func TestCoreDocument_Attribute(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
	cd, err := NewCoreDocument(nil, cas, nil)
	assert.NoError(t, err)
	cd.Attributes = nil
	label := "com.basf.deliverynote.chemicalnumber"
	value := "100"
	key, err := AttrKeyFromLabel(label)
	assert.NoError(t, err)

	// failed get
	assert.False(t, cd.AttributeExists(key))
	_, err = cd.GetAttribute(key)
	assert.Error(t, err)

	// failed delete
	_, err = cd.DeleteAttribute(key, true, nil)
	assert.Error(t, err)

	// success
	attr, err := NewStringAttribute(label, AttrString, value)
	assert.NoError(t, err)
	cd, err = cd.AddAttributes(CollaboratorsAccess{}, true, nil, attr)
	assert.NoError(t, err)
	assert.Len(t, cd.Attributes, 1)
	assert.Len(t, cd.GetAttributes(), 1)

	// check
	assert.True(t, cd.AttributeExists(key))
	attr, err = cd.GetAttribute(key)
	assert.NoError(t, err)
	assert.Equal(t, key, attr.Key)
	assert.Equal(t, label, attr.KeyLabel)
	str, err := attr.Value.String()
	assert.NoError(t, err)
	assert.Equal(t, value, str)
	assert.Equal(t, AttrString, attr.Value.Type)

	// update
	nvalue := "2000"
	attr, err = NewStringAttribute(label, AttrDecimal, nvalue)
	assert.NoError(t, err)
	cd, err = cd.AddAttributes(CollaboratorsAccess{}, true, nil, attr)
	assert.NoError(t, err)
	assert.True(t, cd.AttributeExists(key))
	attr, err = cd.GetAttribute(key)
	assert.NoError(t, err)
	assert.Len(t, cd.Attributes, 1)
	assert.Len(t, cd.GetAttributes(), 1)
	assert.Equal(t, key, attr.Key)
	assert.Equal(t, label, attr.KeyLabel)
	str, err = attr.Value.String()
	assert.NoError(t, err)
	assert.NotEqual(t, value, str)
	assert.Equal(t, nvalue, str)
	assert.Equal(t, AttrDecimal, attr.Value.Type)

	// delete
	cd, err = cd.DeleteAttribute(key, true, nil)
	assert.NoError(t, err)
	assert.Len(t, cd.Attributes, 0)
	assert.Len(t, cd.GetAttributes(), 0)
	assert.False(t, cd.AttributeExists(key))
}

func TestCoreDocument_SetUsedAnchorRepoAddress(t *testing.T) {
	addr := testingidentity.GenerateRandomDID()
	cd := new(CoreDocument)
	cd.SetUsedAnchorRepoAddress(addr.ToAddress())
	assert.Equal(t, addr.ToAddress().Bytes(), cd.AnchorRepoAddress().Bytes())
}

func TestCoreDocument_UpdateAttributes_both(t *testing.T) {
	oldCAttrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	updates := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().Add(60 * time.Hour).UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "new string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("new bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000002",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000002",
		},

		"decimal_test_1": {
			Type:  AttrDecimal.String(),
			Value: "1111.00012",
		},
	}

	oldAttrs := toAttrsMap(t, oldCAttrs)
	newAttrs := toAttrsMap(t, updates)

	newPattrs, err := toProtocolAttributes(newAttrs)
	assert.NoError(t, err)

	oldPattrs, err := toProtocolAttributes(oldAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(oldPattrs, newAttrs)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, newPattrs)
	assert.Equal(t, newAttrs, uattrs)

	oldPattrs[0].Key = utils.RandomSlice(33)
	_, _, err = updateAttributes(oldPattrs, newAttrs)
	assert.Error(t, err)
}

func TestCoreDocument_UpdateAttributes_old_nil(t *testing.T) {
	updates := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().Add(60 * time.Hour).UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "new string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("new bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000002",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000002",
		},

		"decimal_test_1": {
			Type:  AttrDecimal.String(),
			Value: "1111.00012",
		},
	}

	newAttrs := toAttrsMap(t, updates)
	newPattrs, err := toProtocolAttributes(newAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(nil, newAttrs)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, newPattrs)
	assert.Equal(t, newAttrs, uattrs)
}

func TestCoreDocument_UpdateAttributes_updates_nil(t *testing.T) {
	oldCAttrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},
	}

	oldAttrs := toAttrsMap(t, oldCAttrs)
	oldPattrs, err := toProtocolAttributes(oldAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(oldPattrs, nil)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, oldPattrs)
	assert.Equal(t, oldAttrs, uattrs)
}

func TestCoreDocument_UpdateAttributes_both_nil(t *testing.T) {
	upattrs, uattrs, err := updateAttributes(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, upattrs, 0)
	assert.Len(t, uattrs, 0)
}

func TestCoreDocument_Status(t *testing.T) {
	cd, err := NewCoreDocument(nil, CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, cd.GetStatus(), Pending)

	// set status to Committed
	err = cd.SetStatus(Committed)
	assert.NoError(t, err)
	assert.Equal(t, cd.GetStatus(), Committed)

	// try to update status to Committing
	err = cd.SetStatus(Committing)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrCDStatus, err))
	assert.Equal(t, cd.GetStatus(), Committed)
}

func TestCoreDocument_RemoveCollaborators(t *testing.T) {
	did1 := testingidentity.GenerateRandomDID()
	did2 := testingidentity.GenerateRandomDID()
	did3 := testingidentity.GenerateRandomDID() // missing
	cd, err := NewCoreDocument(
		nil,
		CollaboratorsAccess{
			ReadWriteCollaborators: []identity.DID{did1, did},
			ReadCollaborators:      []identity.DID{did1, did2}},
		nil)
	assert.NoError(t, err)
	assert.NoError(t, cd.RemoveCollaborators([]identity.DID{did1}))
	found, err := cd.IsDIDCollaborator(did1)
	assert.NoError(t, err)
	assert.False(t, found)

	found, err = cd.IsDIDCollaborator(did3)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestCoreDocument_AddRole(t *testing.T) {
	key := hexutil.Encode(utils.RandomSlice(32))
	tests := []struct {
		key     string
		collabs []identity.DID
		roleKey []byte
		err     error
	}{
		// empty string
		{
			err: ErrEmptyRoleKey,
		},

		// 30 byte hex
		{
			key:     hexutil.Encode(utils.RandomSlice(30)),
			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
		},

		// random string
		{
			key:     "role key 1",
			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
		},

		// missing collabs
		{
			key: hexutil.Encode(utils.RandomSlice(32)),
			err: ErrEmptyCollaborators,
		},

		// 32 byte key
		{
			key:     key,
			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
		},

		// role exists
		{
			key:     key,
			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
			err:     ErrRoleExist,
		},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	for _, c := range tests {
		r, err := cd.AddRole(c.key, c.collabs)
		if err != nil {
			assert.Equal(t, err, c.err)
			continue
		}

		assert.NoError(t, err)
		assert.Len(t, r.RoleKey, idSize)
	}
}

func TestCoreDocument_UpdateRole(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// invalid role key
	key := utils.RandomSlice(30)
	collabs := []identity.DID{testingidentity.GenerateRandomDID()}
	_, err = cd.UpdateRole(key, collabs)
	assert.Error(t, err)
	assert.Equal(t, err, ErrInvalidRoleKey)

	// missing role
	key = utils.RandomSlice(32)
	_, err = cd.UpdateRole(key, collabs)
	assert.Error(t, err)
	assert.Equal(t, err, ErrRoleNotExist)

	// empty collabs
	r, err := cd.AddRole(hexutil.Encode(key), []identity.DID{testingidentity.GenerateRandomDID()})
	assert.NoError(t, err)
	assert.Equal(t, r.RoleKey, key)
	assert.Len(t, r.Collaborators, 1)
	assert.NotEqual(t, r.Collaborators[0], collabs[0][:])
	_, err = cd.UpdateRole(key, nil)
	assert.Error(t, err)
	assert.Equal(t, err, ErrEmptyCollaborators)

	// success
	r, err = cd.UpdateRole(key, collabs)
	assert.NoError(t, err)
	assert.Equal(t, r.RoleKey, key)
	assert.Len(t, r.Collaborators, 1)
	assert.Equal(t, r.Collaborators[0], collabs[0][:])
	sr, err := cd.GetRole(key)
	assert.NoError(t, err)
	assert.Equal(t, r, sr)
}

func TestFingerprintGeneration(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	cd, err := NewCoreDocument([]byte("inv"), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	role, err := cd.AddRole("test_r", []identity.DID{id1})
	assert.NoError(t, err)
	cd.Document.Roles = append(cd.Document.Roles, role)
	assert.NoError(t, err)
	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)

	// copy over transition rules and roles to generate fingerprint
	f := coredocumentpb.TransitionRulesFingerprint{}
	f.Roles = cd.Document.Roles
	f.TransitionRules = cd.Document.TransitionRules
	p, err := cd.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)

	// create second document with same roles and transition rules to check if generated fingerprint is the same
	cd1, err := NewCoreDocument([]byte("inv"), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	cd1.Document.Roles = cd.Document.Roles
	cd1.Document.TransitionRules = cd.Document.TransitionRules

	f1 := coredocumentpb.TransitionRulesFingerprint{}
	f1.Roles = cd1.Document.Roles
	f1.TransitionRules = cd1.Document.TransitionRules
	p1, err := cd1.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(p, p1))
}
