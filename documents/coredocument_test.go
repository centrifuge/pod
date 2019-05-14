// +build unit

package documents

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
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
		jobsv1.Bootstrapper{},
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
	at := &documentpb.AccessTokenParams{
		Grantee:            "random string",
		DocumentIdentifier: id,
	}
	ncd, err := NewCoreDocumentWithAccessToken(ctxh, CompactProperties("inv"), *at)
	assert.Error(t, err)

	// wrong docID
	at = &documentpb.AccessTokenParams{
		Grantee:            did1.String(),
		DocumentIdentifier: "random string",
	}
	ncd, err = NewCoreDocumentWithAccessToken(ctxh, CompactProperties("inv"), *at)
	assert.Error(t, err)

	// correct access token params
	at = &documentpb.AccessTokenParams{
		Grantee:            did1.String(),
		DocumentIdentifier: id,
	}
	ncd, err = NewCoreDocumentWithAccessToken(ctxh, CompactProperties("inv"), *at)
	assert.NoError(t, err)

	token := ncd.Document.AccessTokens[0]
	assert.Equal(t, token.DocumentIdentifier, cd.Document.DocumentIdentifier)
	assert.Equal(t, token.Grantee, did1[:])
	assert.NotEqual(t, cd.Document.DocumentIdentifier, ncd.Document.DocumentIdentifier)
}

func TestCoreDocument_PrepareNewVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	h := sha256.New()
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
	h = sha256.New()
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
	dr := utils.RandomSlice(32)
	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, dr)
	assert.Nil(t, err)

	cd.GetTestCoreDocWithReset()
	docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, dr)
	assert.Nil(t, err)

	signatureTree, err := cd.getSignatureDataTree()
	assert.Nil(t, err)

	valid, err := proofs.ValidateProofSortedHashes(signingRoot, [][]byte{signatureTree.RootHash()}, docRoot, sha256.New())
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
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), []byte{0, 0, 0, 1}...), lengthLeaf.Property.CompactName())

	signerKey := hexutil.Encode(sig.SignatureId)
	_, signerLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signerLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey), signerLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, append(sig.SignatureId, []byte{0, 0, 0, 2}...)...)...), signerLeaf.Property.CompactName())
	assert.Equal(t, sig.SignerId, signerLeaf.Value)
}

func TestGetDocumentSigningTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// no data root
	_, err = cd.signingRootTree(documenttypes.InvoiceDataTypeUrl, nil)
	assert.Error(t, err)

	// successful tree generation
	tree, err := cd.signingRootTree(documenttypes.InvoiceDataTypeUrl, utils.RandomSlice(32))
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
	dr := utils.RandomSlice(32)

	// successful document root generation
	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, dr)
	assert.NoError(t, err)
	tree, err := cd.DocumentRootTree(documenttypes.InvoiceDataTypeUrl, dr)
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
	h := sha256.New()
	cd, err := newCoreDocument()
	assert.NoError(t, err)
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
	sig := &coredocumentpb.Signature{
		SignerId:    utils.RandomSlice(identity.DIDLength),
		PublicKey:   utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(52),
		Signature:   utils.RandomSlice(32),
	}
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{sig}
	_, err = cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.RootHash())
	assert.NoError(t, err)
	_, err = cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.RootHash())
	assert.NoError(t, err)

	signerKey := hexutil.Encode(sig.SignatureId)
	signProp := fmt.Sprintf("%s.signatures[%s].signature", SignaturesTreePrefix, signerKey)

	cdTree, err := cd.coredocTree(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	signTree, err := cd.getSignatureDataTree()
	tests := []struct {
		fieldName    string
		fromCoreDoc  bool
		fromSignTree bool
		proofLength  int
	}{
		{
			"prefix.sample_field",
			false,
			false,
			3,
		},
		{
			CDTreePrefix + ".document_identifier",
			true,
			false,
			6,
		},
		{
			"prefix.sample_field2",
			false,
			false,
			3,
		},
		{
			CDTreePrefix + ".next_version",
			true,
			false,
			6,
		},
		{
			signProp,
			false,
			true,
			4,
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
			} else if test.fromSignTree {
				_, l = signTree.GetLeafByProperty(test.fieldName)
				valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes[:3], signTree.RootHash(), h)
				assert.NoError(t, err)
				assert.True(t, valid)
			} else {
				_, l = testTree.GetLeafByProperty(test.fieldName)
				assert.Contains(t, compactProps, l.Property.CompactName())
				valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes[:1], testTree.RootHash(), h)
				assert.NoError(t, err)
				assert.True(t, valid)
			}
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes, docRoot, h)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}
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
	_, err = cd.DeleteAttribute(key)
	assert.Error(t, err)

	// success
	attr, err := NewAttribute(label, AttrString, value)
	assert.NoError(t, err)
	cd, err = cd.AddAttributes(attr)
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
	attr, err = NewAttribute(label, AttrDecimal, nvalue)
	assert.NoError(t, err)
	cd, err = cd.AddAttributes(attr)
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
	cd, err = cd.DeleteAttribute(key)
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
	oldCAttrs := map[string]*documentpb.Attribute{
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

	updates := map[string]*documentpb.Attribute{
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

	oldAttrs, err := FromClientAttributes(oldCAttrs)
	assert.NoError(t, err)

	newAttrs, err := FromClientAttributes(updates)
	assert.NoError(t, err)

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
	updates := map[string]*documentpb.Attribute{
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

	newAttrs, err := FromClientAttributes(updates)
	assert.NoError(t, err)

	newPattrs, err := toProtocolAttributes(newAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(nil, newAttrs)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, newPattrs)
	assert.Equal(t, newAttrs, uattrs)
}

func TestCoreDocument_UpdateAttributes_updates_nil(t *testing.T) {
	oldCAttrs := map[string]*documentpb.Attribute{
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

	oldAttrs, err := FromClientAttributes(oldCAttrs)
	assert.NoError(t, err)

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

func TestValidateFullNFTProof(t *testing.T) {
	payload := `{
  "header": {
    "document_id": "0xa7f7933dcb42715924b4c06f425238f324fb499e6128c66045592b0258f38c23",
    "version_id": "0xf138a99b8ee613f96704d9d2a49a1dfad483ab8ec93ac768d47775d9bda08d74",
		"document_root": "0xfd951b063ce425a62f2ab49b20b3228194fcf7d31f2bb4851cbb6659d5a87654"
  },
  "field_proofs": [
    {
      "property": "0x000100000000000e",
      "value": "0x0006aaf7c8516d0c0000",
      "salt": "0x6c840dacf1b0139b18b2fedacef55efddb73ae37e9db0a19e9a1709bca9d4f1c",
      "hash": "0x",
      "sorted_hashes": [
        "0xdab2a11ac7e4ddb1492de65d7bb45c133330c4c593caff732ec962790ac286b6",
        "0xca11c2963755ad58ed50ae009f5b65abee0fbf10d01066edc0ba209c899d5b07",
        "0x2b690962582ab53ba24eddd3134e3d455b4ede653fa7885688e618a1d41dbef6",
        "0x331aafe3fa23bbbf4bfcd68e2abcf4b3c37f5970dfa7431dbf72905047bb9668",
        "0x2e0a9a75c554e885167857f998fb41e420f5e7335cfdb39e2e56de4e2fce2a66",
        "0x98824b5434b1d6043ac3b2d7537828ce40ddae346c5614ced4b704b4ac39279d",
        "0xa58a65a70a2964dd264c5e57c3b4260044d2775372e6d9356499731e7029e6ca",
        "0x960a4e462fe4fa1bffaacf1877042527e757445cce8e317ec227518b5659b3b6",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x000100000000000d",
      "value": "0x455552",
      "salt": "0x0419fab213d33acbe02bc205f14752b6320218f26eafcd56e4914e44553991a3",
      "hash": "0x",
      "sorted_hashes": [
        "0x32ddd71157dc2f72864a54c53fb502bcfbe6cf53cd75f160d75e447da968ebd1",
        "0xca11c2963755ad58ed50ae009f5b65abee0fbf10d01066edc0ba209c899d5b07",
        "0x2b690962582ab53ba24eddd3134e3d455b4ede653fa7885688e618a1d41dbef6",
        "0x331aafe3fa23bbbf4bfcd68e2abcf4b3c37f5970dfa7431dbf72905047bb9668",
        "0x2e0a9a75c554e885167857f998fb41e420f5e7335cfdb39e2e56de4e2fce2a66",
        "0x98824b5434b1d6043ac3b2d7537828ce40ddae346c5614ced4b704b4ac39279d",
        "0xa58a65a70a2964dd264c5e57c3b4260044d2775372e6d9356499731e7029e6ca",
        "0x960a4e462fe4fa1bffaacf1877042527e757445cce8e317ec227518b5659b3b6",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x0001000000000016",
      "value": "0x000000005cdec1e8",
      "salt": "0x46ee6bef884efc23882b166c65e6c1e91db7f8466b053cc4a233847b670478ef",
      "hash": "0x",
      "sorted_hashes": [
        "0xdac4d2f0fd4197da1f269eacc84a963901d371fcb840bcaa8082ca88bb9b64a7",
        "0x9e32aaea1982be17541dccf73380a1c255523ab88f274775fe809c43859bc304",
        "0x42637f76c88806ae7956d79727f432f25218b3d50b177bc5ce9bf457ff19a2b1",
        "0xbeb657bc68986e547a89476b06efba7c12ba2e13886d79345757c848e1773ff0",
        "0xf412661e5f43da2a32e8d4c7115bfe264dd71130ba0a60a218cf11edc1850256",
        "0x98824b5434b1d6043ac3b2d7537828ce40ddae346c5614ced4b704b4ac39279d",
        "0xa58a65a70a2964dd264c5e57c3b4260044d2775372e6d9356499731e7029e6ca",
        "0x960a4e462fe4fa1bffaacf1877042527e757445cce8e317ec227518b5659b3b6",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x0001000000000013",
      "value": "0xc87e3bff78437cf622f5f6eb37afb6cb153c6cb1",
      "salt": "0xb2806c8dea19bb0c3c6ba2e2e23fdc32b5ec477dc304cdde47124f95830e6903",
      "hash": "0x",
      "sorted_hashes": [
        "0x268ea0ae7a275cbc5b0b18d0fd4ea10654e6a8dfcefee6cdd2095779ad897795",
        "0xf988a8a4b588dd1759c59ac2acfca18becee3a8f11411baace107088e6fda7ac",
        "0x335edc636da6fb40af07e193687359699f42e2409383418b79edd2c210b02f82",
        "0xbeb657bc68986e547a89476b06efba7c12ba2e13886d79345757c848e1773ff0",
        "0xf412661e5f43da2a32e8d4c7115bfe264dd71130ba0a60a218cf11edc1850256",
        "0x98824b5434b1d6043ac3b2d7537828ce40ddae346c5614ced4b704b4ac39279d",
        "0xa58a65a70a2964dd264c5e57c3b4260044d2775372e6d9356499731e7029e6ca",
        "0x960a4e462fe4fa1bffaacf1877042527e757445cce8e317ec227518b5659b3b6",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x0001000000000002",
      "value": "0x756e70616964",
      "salt": "0x4080d50ced302ba1a81a24a72f51a2d086afc2c4c3aaec6ece3b4ecb0d9847ec",
      "hash": "0x",
      "sorted_hashes": [
        "0xc92d9af97831300bb37d4ea586fc636a9c81be5a02bc886d90bdd5a3c432bee3",
        "0xd7737309f59deb941d9ba401160d25f0fd7f1d530e9a4fb16e47bf56a24823c2",
        "0x40975f0e99d3036c8b64c3c48c625b210654a641cc00c69ec300814777bba505",
        "0xaf915a2e1ef8ab6afb01a8f676068e5fee91e56eb4674494ec4233662daf75ff",
        "0x2e0a9a75c554e885167857f998fb41e420f5e7335cfdb39e2e56de4e2fce2a66",
        "0x98824b5434b1d6043ac3b2d7537828ce40ddae346c5614ced4b704b4ac39279d",
        "0xa58a65a70a2964dd264c5e57c3b4260044d2775372e6d9356499731e7029e6ca",
        "0x960a4e462fe4fa1bffaacf1877042527e757445cce8e317ec227518b5659b3b6",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x040000000000000a",
      "value": "0x",
      "salt": "0x",
      "hash": "0xf548b47a8f8f987465206212f781ba13e9348c43ccb5a122fc224db5a23af4fc",
      "sorted_hashes": [
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x0300000000000001c87e3bff78437cf622f5f6eb37afb6cb153c6cb1000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c500000004",
      "value": "0x2b795d0a2bcad9a4ff4b275b0de9d73fdaf081ca81828e9777d3c0c9cac1ce4f7940272d1a8ce3e52e84606552acd7e7b19cd674e21e59e9163d0d11d50acea200",
      "salt": "0xa929394ddb08f432541c75cd4c527f28eedafa5172ba1c0fa23bcbaeddd2b646",
      "hash": "0x",
      "sorted_hashes": [
        "0x649cfc4e6ff0ed826a813d270eb496ffca3e887daef71ff93817f619f7ac4281",
        "0x7daece4242f41a77a2627829e087dd2333035efc8313b591895d30f718bfcd0c",
        "0xdcabd8a68d16fb86e5693ff833d7ba55446381bbb536b6609b180c683702c4af",
        "0xf548b47a8f8f987465206212f781ba13e9348c43ccb5a122fc224db5a23af4fc"
      ]
    },
    {
      "property": "0x0300000000000001c87e3bff78437cf622f5f6eb37afb6cb153c6cb1000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c500000005",
      "value": "0x00",
      "salt": "0x6336a76b92c19d8348365c52a4e881847d014800aca5452fdfbec7bb8bd92bfd",
      "hash": "0x",
      "sorted_hashes": [
        "0x8cab378ea94fe7e4ede501ec98f9f47afc30f17bd3e90328b461cb0f147f0147",
        "0xf548b47a8f8f987465206212f781ba13e9348c43ccb5a122fc224db5a23af4fc"
      ]
    },
    {
      "property": "0x0100000000000004",
      "value": "0xe1f58d2e95b541277ada1f0a1666abb48b7f7cbe47785d47db76d4e3755cde43",
      "salt": "0x29c14c71251aa9f2ec3af09b0178df85b4d2c62d8e532c2fad810c4d685aa8bc",
      "hash": "0x",
      "sorted_hashes": [
        "0xe60424c2ac201148bf9c4b9b6e1d780233a31b999c0b02e1b0f71bb8e27469cd",
        "0x475d0efa801402c605aa1c3201991c9654aa36b84f736100f37449ad6dcc6c84",
        "0x53b0689ca95922cb512c37bc82e678826a45ed54d22a709b9ec4515e242ea6fa",
        "0x84432ce7326da0608f4bafae202ab728c890d73068eb6476b2dce086df760c77",
        "0x8446e23d5571be38b412849b43b2c3eaa33ebfbfee0412e24c4f30709d482b9b",
        "0x384d38cb9f0f632ad738b99cd77f5a38535b24ef54ee6033a2ae0b328629fa33",
        "0x25d7811b737d549feb402716493dd71c37f4cee7cc185f6fe872a96f8ce99cac",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x01000000000000144e498bd371833d6b910513a6e15f1bd8e97fbeda000000000000000000000000",
      "value": "0x385c967913dcd80d6867b76d6d4b6cc6790da8bfa4579c665a26f579d0ef8c4f",
      "salt": "0x117fe843ca6e723699c6a434a408ec1144c9cec7062d06ccb1cecb7c54a11d4c",
      "hash": "0x",
      "sorted_hashes": [
        "0x7d6250d44d579ba602b87c918405c3844f45817456ae1f6b23e67969815a1d1f",
        "0xf08a52f41b24f8d26e6ed36eacf1468a00248cf9b74b3c96b4bb242bdcb0131e",
        "0x109b43a43276c3a51f1c91c202974bf964acfc2168c8f63833fd63a659a1c843",
        "0x552fc6b4c3f8ca4ccf7a1cd7de266231ac6228391ea942197bdd71a25715bead",
        "0x2372b6a9d99e27e484f8088ab64a1acb9227497a366d967adff86e377e956006",
        "0x384d38cb9f0f632ad738b99cd77f5a38535b24ef54ee6033a2ae0b328629fa33",
        "0x25d7811b737d549feb402716493dd71c37f4cee7cc185f6fe872a96f8ce99cac",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x01000000000000130000000000000001000000020000000000000000",
      "value": "0x81504b56a32fdd0561f8532393739fb2e8bd74c4796b3d6872f86d26acfdbc40",
      "salt": "0x0b79dce578e35e5e0a929e4d63145752fb7f97c5386b00662a093bac6ba88975",
      "hash": "0x",
      "sorted_hashes": [
        "0xff1d476a627a381d6c0cb181bef9d910263f7a115328967f3ac6efb8e2ea5bb2",
        "0x60192baba91c10be6efd85c83de332d49d6601ba39742f5ff58a4f05a846b6e6",
        "0x4176ff6b2b8a834ef2402b4d4247d45564ffeb3fd27423b797775855050e0d86",
        "0x552fc6b4c3f8ca4ccf7a1cd7de266231ac6228391ea942197bdd71a25715bead",
        "0x2372b6a9d99e27e484f8088ab64a1acb9227497a366d967adff86e377e956006",
        "0x384d38cb9f0f632ad738b99cd77f5a38535b24ef54ee6033a2ae0b328629fa33",
        "0x25d7811b737d549feb402716493dd71c37f4cee7cc185f6fe872a96f8ce99cac",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x0100000000000013000000000000000100000004",
      "value": "0x0000000000000002",
      "salt": "0xff7df0e49d0c2fd222bf8aecd7ab9f1859a398b0b85ea97c90cc7fc521d6d54f",
      "hash": "0x",
      "sorted_hashes": [
        "0x5dd00536662f9c1255c5e7eb3bfed44106f40399c61ae7c002ae829ba159fa57",
        "0x26d35d5232b00e05162061f5a571727ab3e8f68df697c66ea3b49e87d4f10540",
        "0x109b43a43276c3a51f1c91c202974bf964acfc2168c8f63833fd63a659a1c843",
        "0x552fc6b4c3f8ca4ccf7a1cd7de266231ac6228391ea942197bdd71a25715bead",
        "0x2372b6a9d99e27e484f8088ab64a1acb9227497a366d967adff86e377e956006",
        "0x384d38cb9f0f632ad738b99cd77f5a38535b24ef54ee6033a2ae0b328629fa33",
        "0x25d7811b737d549feb402716493dd71c37f4cee7cc185f6fe872a96f8ce99cac",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    },
    {
      "property": "0x010000000000000181504b56a32fdd0561f8532393739fb2e8bd74c4796b3d6872f86d26acfdbc40000000040000000000000000",
      "value": "0x4e498bd371833d6b910513a6e15f1bd8e97fbeda385c967913dcd80d6867b76d6d4b6cc6790da8bfa4579c665a26f579d0ef8c4f",
      "salt": "0x0a6867a026a4355c78f3091d69cb04a74c4f892ffaa982ab4e866a28deba14b0",
      "hash": "0x",
      "sorted_hashes": [
        "0x09f0fe3f43dd56934cb2d6b6d9632f1de3f6189d9edaf79ae078e2cdbe9840be",
        "0xc937c961fb9baa405ea2e3843cceeccdf95e267c46704dc8ffe0d24086492db5",
        "0x95866dc19012b7134392220242bd11c4cf87b309764936653a6bd383c71e7bdb",
        "0xfb7e2e1840d544342f65ba3bad6dd9fa1b6e7d3918c14f26a5228091067083be",
        "0x8446e23d5571be38b412849b43b2c3eaa33ebfbfee0412e24c4f30709d482b9b",
        "0x384d38cb9f0f632ad738b99cd77f5a38535b24ef54ee6033a2ae0b328629fa33",
        "0x25d7811b737d549feb402716493dd71c37f4cee7cc185f6fe872a96f8ce99cac",
        "0x3e0ff76ad9759a84927bc0df2821b6a5091f32dd3a61db3f75082f71d63a6ad2"
      ]
    }
  ]
}`

	type Header struct {
		DocumentId   string `json:"document_id"`
		VersionId    string `json:"version_id"`
		DocumentRoot string `json:"document_root"`
	}

	type FieldProof struct {
		Property     string   `json:"property"`
		Value        string   `json:"value"`
		Salt         string   `json:"salt"`
		Hash         string   `json:"hash"`
		SortedHashes []string `json:"sorted_hashes"`
	}

	type Payload struct {
		Header      Header       `json:"header"`
		FieldProofs []FieldProof `json:"field_proofs"`
	}

	var obj Payload
	err := json.Unmarshal([]byte(payload), &obj)
	assert.NoError(t, err)

	for i := 0; i < len(obj.FieldProofs); i++ {
		var lh []byte
		if obj.FieldProofs[i].Hash == "0x" {
			prop, err := hexutil.Decode(obj.FieldProofs[i].Property)
			assert.NoError(t, err)
			val, err := hexutil.Decode(obj.FieldProofs[i].Value)
			assert.NoError(t, err)
			salt, err := hexutil.Decode(obj.FieldProofs[i].Salt)
			assert.NoError(t, err)
			lh, err = crypto.Sha256Hash(append(prop, append(val, salt...)...))
			assert.NoError(t, err)
		} else {
			lh, err = hexutil.Decode(obj.FieldProofs[i].Hash)
			assert.NoError(t, err)
		}
		var sh [][]byte
		for j := 0; j < len(obj.FieldProofs[i].SortedHashes); j++ {
			shi, err := hexutil.Decode(obj.FieldProofs[i].SortedHashes[j])
			assert.NoError(t, err)
			sh = append(sh, shi)
		}
		rh, err := hexutil.Decode(obj.Header.DocumentRoot)
		assert.NoError(t, err)
		valid, err := proofs.ValidateProofSortedHashes(lh, sh, rh, sha256.New())
		assert.NoError(t, err)
		msg := fmt.Sprintf("Failed for proof %d", i)
		assert.True(t, valid, msg)
	}

}
