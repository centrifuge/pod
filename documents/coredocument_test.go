// +build unit

package documents

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/centrifuge/go-centrifuge/crypto"
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
			fmt.Println(p)
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
    "document_id": "0xe979d2bc11a85dc51f7e4c8bdfd791ba5608101e3c714a872471ddf1e51f6ff6",
    "version_id": "0xba9daa7ba26b15f1d26991ef7e7da3503f618218e7c6e4ce36a2d930845827b5",
    "document_root": "0x900b0aa1f09bd045a37ed8bf48e38f28ac0ea3661ee3b784a12ca5067f91709d"
  },
  "field_proofs": [
    {
      "property": "0x000100000000000e",
      "value": "0x0006aaf7c8516d0c0000",
      "salt": "0x785edfe16b979ad72e3a501e101a29315caa1999b14330c421ef1f6d0d1cf4c0",
      "hash": "0x",
      "sorted_hashes": [
        "0x1d88d415b3615ac4e6862cfc777039d59ac77f11d1c186db5376f8a3069692c4",
        "0x3bbbbd508480b46d438bac382096212f2bdc1fd02b5dcce952ea762c9635ee09",
        "0xc17e2aea5059219d3a99b636716af75bcbf39634dc8e594a0e2088bdab9ac441",
        "0xada3d2717a2c816310bae031cd68b48b49e09c9a5cbab56377b188acdfa07933",
        "0xfc20d3cbdfdc02358ce25c0856dd4c8c1509a023d883c2566964bd5cb9b17aa7",
        "0xd6d1756761e9af6c09b35393b2a6b4e8a658db72d5f3f63edf4683632669bbfe",
        "0xf9e5f0be31bd9e6e77872f4b44b05607d0984b6a6591e60c9b0a4e7748a23947",
        "0x6c10c51e9b1184ab848892e660c84134a4aa7902924d51329cdfca84cbd4ea43"
      ]
    },
    {
      "property": "0x000100000000000d",
      "value": "0x455552",
      "salt": "0xf2d303c7ec7e2c92536df61a80b05456a0a00fcd7531950352188d87e1a6b8b0",
      "hash": "0x",
      "sorted_hashes": [
        "0x8c9fd3039aed36c890e37b23c59d97e8de33c8f197a2ec9ed4ab28ac329f18ed",
        "0x3bbbbd508480b46d438bac382096212f2bdc1fd02b5dcce952ea762c9635ee09",
        "0xc17e2aea5059219d3a99b636716af75bcbf39634dc8e594a0e2088bdab9ac441",
        "0xada3d2717a2c816310bae031cd68b48b49e09c9a5cbab56377b188acdfa07933",
        "0xfc20d3cbdfdc02358ce25c0856dd4c8c1509a023d883c2566964bd5cb9b17aa7",
        "0xd6d1756761e9af6c09b35393b2a6b4e8a658db72d5f3f63edf4683632669bbfe",
        "0xf9e5f0be31bd9e6e77872f4b44b05607d0984b6a6591e60c9b0a4e7748a23947",
        "0x6c10c51e9b1184ab848892e660c84134a4aa7902924d51329cdfca84cbd4ea43"
      ]
    },
    {
      "property": "0x0001000000000016",
      "value": "0x000000005cdc17e4",
      "salt": "0x7668a87b1c9a9263caf506c9a21dcf8a29827d6cb0f60600c200130aab14fb34",
      "hash": "0x",
      "sorted_hashes": [
        "0x97267b0caf7240b7d40a6086555e70667c3886ec479d556df8b1c7a45acbe55f",
        "0x804df462e23e608e2b2ec07b70b7c3858a1695f0107471640e72fe0ba5f27b3f",
        "0xbdd88484d8043570833b6b6975db6e1b2cd6734957a74394320cc74250fb5f70",
        "0xed8bc9e5f9b05a9a3394bfcc1b4e6b73a8fcc62462bd0650e927656805ec8b76",
        "0x5ed15d971bb89862b026bbe7e14b758319ea8ee185e6901e451a98859083462d",
        "0xd6d1756761e9af6c09b35393b2a6b4e8a658db72d5f3f63edf4683632669bbfe",
        "0xf9e5f0be31bd9e6e77872f4b44b05607d0984b6a6591e60c9b0a4e7748a23947",
        "0x6c10c51e9b1184ab848892e660c84134a4aa7902924d51329cdfca84cbd4ea43"
      ]
    },
    {
      "property": "0x0001000000000013",
      "value": "0x4118ff654d365dab7dbfade7613b3e9643cf4197",
      "salt": "0xaa905d6d1623ad7c29e678d5b0e86ac8a6f089c5dabed99d01128c2dddfab73e",
      "hash": "0x",
      "sorted_hashes": [
        "0xdd73ebc09d64f9afe115e7389318b633783cbfeccf9ccb8d4d25a4bd4383399d",
        "0xe26a8639a9b2e3d79101e28e2403caaded732d7c9e9239569e61268401e5098f",
        "0x161fe5e30fd341c73e0192edbb6ed10013d3ca436e8de5c7305f01db704b9b07",
        "0xed8bc9e5f9b05a9a3394bfcc1b4e6b73a8fcc62462bd0650e927656805ec8b76",
        "0x5ed15d971bb89862b026bbe7e14b758319ea8ee185e6901e451a98859083462d",
        "0xd6d1756761e9af6c09b35393b2a6b4e8a658db72d5f3f63edf4683632669bbfe",
        "0xf9e5f0be31bd9e6e77872f4b44b05607d0984b6a6591e60c9b0a4e7748a23947",
        "0x6c10c51e9b1184ab848892e660c84134a4aa7902924d51329cdfca84cbd4ea43"
      ]
    },
    {
      "property": "0x0001000000000002",
      "value": "0x756e70616964",
      "salt": "0xa3a3bded8532d78bdacff602d10000e481176aaa17ada0bd953893ec138786e6",
      "hash": "0x",
      "sorted_hashes": [
        "0x50137b36675fab13eff889544459560721ae5b25de6b597e3c812a1702093771",
        "0xfe6b162a08a9769c725dd0a5868d28a3978b277a00def8ddecfecddeb8a45810",
        "0x1fc9417d68715edbbf94bfa1fd74e6a29601184dbbd55a940640bdb5d5db5781",
        "0x0640a4c24b3e953eabb83e53534b49f648a9a025c218d80425114b4d77ac47a8",
        "0xfc20d3cbdfdc02358ce25c0856dd4c8c1509a023d883c2566964bd5cb9b17aa7",
        "0xd6d1756761e9af6c09b35393b2a6b4e8a658db72d5f3f63edf4683632669bbfe",
        "0xf9e5f0be31bd9e6e77872f4b44b05607d0984b6a6591e60c9b0a4e7748a23947",
        "0x6c10c51e9b1184ab848892e660c84134a4aa7902924d51329cdfca84cbd4ea43"
      ]
    },
    {
      "property": "0x040000000000000a",
      "value": "0x",
      "salt": "0x",
      "hash": "0x900b0aa1f09bd045a37ed8bf48e38f28ac0ea3661ee3b784a12ca5067f91709d",
      "sorted_hashes": [
        "0x254687f1881b440315638cbbd360a725a00c8523e63b83858d74cd909bf20ddb"
      ]
    },
    {
      "property": "0x03000000000000014118ff654d365dab7dbfade7613b3e9643cf4197000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c500000004",
      "value": "0xa75078a88d69318f4494351d93c1d1d4048bcb4d51f3d3be7068acdb997417037f6c7f5778e1f1553749b0c6d8f30541aa3042424e26e7f0093aaa4440e1d62d01",
      "salt": "0x54cfd844b2764ed59a5ff92f04ff4a1a482dd3817f66b5d21600225a844d1899",
      "hash": "0x",
      "sorted_hashes": [
        "0xc197324428126f72ea67e94cc500983b4b7ade76a606f224456e3ff700f50827",
        "0x58baae73626b1930813540d9805a6381483c5ed511e07f7c09121e91787350a6",
        "0x18c3daa22c679946fb41ab72fa7da778266a444dc0ab4acbe043fa821b095fbc",
        "0x900b0aa1f09bd045a37ed8bf48e38f28ac0ea3661ee3b784a12ca5067f91709d"
      ]
    },
    {
      "property": "0x03000000000000014118ff654d365dab7dbfade7613b3e9643cf4197000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c500000005",
      "value": "0x00",
      "salt": "0x661a033fa4702a8be8f1ba1feb2866bfff7840a8a56d51f890afc9278c22a2c6",
      "hash": "0x",
      "sorted_hashes": [
        "0xcad760bb8cfea7712d27ce7eb4e04dffd6c7b217351154466fb9cd8d50e8e5c6",
        "0x900b0aa1f09bd045a37ed8bf48e38f28ac0ea3661ee3b784a12ca5067f91709d"
      ]
    },
    {
      "property": "0x0100000000000004",
      "value": "0x5e6d0767c4246d8d82dba594ea7a1676d229f5295dbefe75565b675d838f440e",
      "salt": "0x29996a9f11b42157e52a917c96986e964784101311c2c16ede78dbb962793a83",
      "hash": "0x",
      "sorted_hashes": [
        "0x912ee8ae5b0946e836758b201f2ec70aa06e706ce841bc6db8edd3dfad65624b",
        "0xa2016043175ca41d08a11cbc474aaa7bc5d5d9bc41030a1719531ad1a5304980",
        "0xdba50565b17e0e1eb3b1954792312552d87341bc107ac012eb6830a5986062de",
        "0x48e3564c2c51b02711c4c412c1a7b35e3237e787f8b60fad50cc7559514479f5",
        "0xfc2821ade4d173a1a3889d7aca0088dee539b67dbcd57556ca14e306e31c67ad",
        "0x4d5c80bf42ff002ab5bf8e097ee5135d1c93696be79b614c8f54a50c3353cda1",
        "0x9e8d6e66e54c438b16eaa481ec96b03a5e718706c34a07eb9f8018653530c9b7"
      ]
    },
    {
      "property": "0x01000000000000144e498bd371833d6b910513a6e15f1bd8e97fbeda000000000000000000000000",
      "value": "0xe0f246323d950753a81b6db18691797b1f918ac7a900f08eabef038d4bdffe0d",
      "salt": "0x632dc1882caf0f9e28e60d994b8f49f1c8f309e9d5fee15b195a14659bb2cb22",
      "hash": "0x",
      "sorted_hashes": [
        "0xdaefb4a9c611fe507a0f7abcfdda281870b1081d11e95580b811e54a76ead937",
        "0x59fd237a4c05f90a5651f9a703a83dc40a8df66ea8f73f116c1c848163e04ec0",
        "0x0168a18ab5c6252cfafc741d58b3293bf15d82c088bbe369c02bb191920c18ac",
        "0x378e6213732eaff7f6276158ad02ae38b44f4f45fd9db5300e0b3cdbf1f01dda",
        "0x648fb23dd2f0e72cd90fb682f22f366f23cc003482c4896ea28cb44a482a9070",
        "0x4d5c80bf42ff002ab5bf8e097ee5135d1c93696be79b614c8f54a50c3353cda1",
        "0x9e8d6e66e54c438b16eaa481ec96b03a5e718706c34a07eb9f8018653530c9b7"
      ]
    },
    {
      "property": "0x01000000000000130000000000000001000000020000000000000000",
      "value": "0x83e40a6c67b21dc7c9f8a432f4de394de9a9eaf14d00d98a08707e1ddfe15bf0",
      "salt": "0xf3942868386f1aa33de381ce905e9cc041684c452ad877b9e1d197824b36fd3f",
      "hash": "0x",
      "sorted_hashes": [
        "0x070a5d4061ca24d661bccd46b70ca034b9a1687280f9787597cbecf2d1e85f83",
        "0x0d3b60bf5f7c8bdb95ffe544afec77710df538a502fe5abd39a4dd1edb9f576c",
        "0x58a739f5eac5d7538d91c9b6bb903eafb3df320842e70d744f1dc07023dfac33",
        "0x378e6213732eaff7f6276158ad02ae38b44f4f45fd9db5300e0b3cdbf1f01dda",
        "0x648fb23dd2f0e72cd90fb682f22f366f23cc003482c4896ea28cb44a482a9070",
        "0x4d5c80bf42ff002ab5bf8e097ee5135d1c93696be79b614c8f54a50c3353cda1",
        "0x9e8d6e66e54c438b16eaa481ec96b03a5e718706c34a07eb9f8018653530c9b7"
      ]
    },
    {
      "property": "0x0100000000000013000000000000000100000004",
      "value": "0x0000000000000002",
      "salt": "0xb9c4b338979eeb3d6e22c8e2ea32cfba8deacfd31f850b9be644a8e97109550e",
      "hash": "0x",
      "sorted_hashes": [
        "0xda0f27b5002497547ff6b5c78159d8d75e5539c6d849da0ec585d324d6406dfa",
        "0xd86877f59e511eaf637a0c7d4ed1db9ddbafcc9a161db94432c17c22a5a3153a",
        "0x0168a18ab5c6252cfafc741d58b3293bf15d82c088bbe369c02bb191920c18ac",
        "0x378e6213732eaff7f6276158ad02ae38b44f4f45fd9db5300e0b3cdbf1f01dda",
        "0x648fb23dd2f0e72cd90fb682f22f366f23cc003482c4896ea28cb44a482a9070",
        "0x4d5c80bf42ff002ab5bf8e097ee5135d1c93696be79b614c8f54a50c3353cda1",
        "0x9e8d6e66e54c438b16eaa481ec96b03a5e718706c34a07eb9f8018653530c9b7"
      ]
    },
    {
      "property": "0x010000000000000183e40a6c67b21dc7c9f8a432f4de394de9a9eaf14d00d98a08707e1ddfe15bf0000000040000000000000000",
      "value": "0x4e498bd371833d6b910513a6e15f1bd8e97fbedae0f246323d950753a81b6db18691797b1f918ac7a900f08eabef038d4bdffe0d",
      "salt": "0x5d69bb3beaf5df8fdfcf7f6be26be2ff82b7b04c44d4e82416947dd9210efb72",
      "hash": "0x",
      "sorted_hashes": [
        "0xe1fded6edbd647d7b29e1f82b42bf067cd420dd71be35a8c7ab8ed55e240c57a",
        "0x93ab99771a9e6685609595dca57137eb57e889c83bb5890b457715c0e995f427",
        "0xe2a1661ed3529459cb225fba7db41f98ffd36e63dadd2d2356bbf8c292c58313",
        "0x71590f9e0fd4554d4c9724f3fa7c465b7e74831c0ba557cd922d74bddd2c78db",
        "0xfc2821ade4d173a1a3889d7aca0088dee539b67dbcd57556ca14e306e31c67ad",
        "0x4d5c80bf42ff002ab5bf8e097ee5135d1c93696be79b614c8f54a50c3353cda1",
        "0x9e8d6e66e54c438b16eaa481ec96b03a5e718706c34a07eb9f8018653530c9b7"
      ]
    }
  ]
}`
	type Header struct {
		DocumentId string `json:"document_id"`
		VersionId string `json:"version_id"`
		DocumentRoot string `json:"document_root"`
	}

	type FieldProof struct {
		Property string `json:"property"`
		Value string `json:"value"`
		Salt string `json:"salt"`
		Hash string `json:"hash"`
		SortedHashes []string `json:"sorted_hashes"`
	}

	type Payload struct {
		Header Header	`json:"header"`
		FieldProofs []FieldProof `json:"field_proofs"`
	}

	var obj Payload
	err := json.Unmarshal([]byte(payload), &obj)
	if err != nil {
		panic(err)
	}

	for i := 0 ; i < len(obj.FieldProofs) ; i++ {
		var lh []byte
		if obj.FieldProofs[i].Hash == "0x" {
			prop, _ := hexutil.Decode(obj.FieldProofs[i].Property)
			val, _ := hexutil.Decode(obj.FieldProofs[i].Value)
			salt, _ := hexutil.Decode(obj.FieldProofs[i].Salt)
			lh, _ = crypto.Sha256Hash(append(prop, append(val, salt...)...))
		} else {
			lh, _ = hexutil.Decode(obj.FieldProofs[i].Hash)
		}
		var sh [][]byte
		for j := 0 ; j < len(obj.FieldProofs[i].SortedHashes); j++ {
			shi, _ := hexutil.Decode(obj.FieldProofs[i].SortedHashes[j])
			sh = append(sh, shi)
		}
		rh, _ := hexutil.Decode(obj.Header.DocumentRoot)
		valid, err := proofs.ValidateProofSortedHashes(lh, sh, rh, sha256.New())
		assert.NoError(t, err)
		msg := fmt.Sprintf("Failed for proof %d", i)
		assert.True(t, valid, msg)
	}

}
