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
    "document_id": "0x9ce93d6e68328b0b673fbe141dba6383cb18a5b0c7bf792881c006338aa783f7",
    "version_id": "0x1c8c7dbe078db3202949f06bc311f85dd3052c376bf129585459e11a4c90b53c",
    "document_root": "0x4dcbbaaf85a0c7543cf9ce918444b654725e6b3a7934a33a8dd5530c3f939055"
  },
  "field_proofs": [
    {
      "property": "0x000100000000000e",
      "value": "0x0006aaf7c8516d0c0000",
      "salt": "0x0492e6cfb43d36bc1ae5ca44290b336c6d8b9b0518c2cccb15f99d84d4f04271",
      "hash": "0x",
      "sorted_hashes": [
        "0x436ccb22888c3d43efd3413564783b2b13c5d1d77b95ba6da30c0485ec92a0ec",
        "0xdd9f41d7ddeec40f932d16baa987fe18252a332b5fe51cbf3dfefcc991413f30",
        "0x7faba8f6d719e19d7215a458455af4e49c092b3dff27a4685aed3a5a9e62cfef",
        "0x5e1fd03d6f23b684bd352c748fff0e801bb895c133b07686840916841699dac3",
        "0xf559ea3a8be0f2eedf214c2a90c40377602cbacc67a970fbced9721220944b55",
        "0xe598fd9ad7fb03bc8a5c34c9b70e5fa8d11777520a70c8c0b762a7e22ea1ab2c",
        "0x82c018f529283075e9d8cdb9c75aca5e1e0064dc335a19a29186aae04a3ea182",
        "0xad9f9ee1bd48c449c9dd7445c79cfb5aad5b9dc678690da848494303d17b8512",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x000100000000000d",
      "value": "0x455552",
      "salt": "0x2bb177e6392bdb40fcc85897776d18e78afe38aa93d4d10c36bde1f35038e34f",
      "hash": "0x",
      "sorted_hashes": [
        "0xb87a691fef57eecefb537962bffb3cc2510aacd5050f35c439c51d78fa388812",
        "0xdd9f41d7ddeec40f932d16baa987fe18252a332b5fe51cbf3dfefcc991413f30",
        "0x7faba8f6d719e19d7215a458455af4e49c092b3dff27a4685aed3a5a9e62cfef",
        "0x5e1fd03d6f23b684bd352c748fff0e801bb895c133b07686840916841699dac3",
        "0xf559ea3a8be0f2eedf214c2a90c40377602cbacc67a970fbced9721220944b55",
        "0xe598fd9ad7fb03bc8a5c34c9b70e5fa8d11777520a70c8c0b762a7e22ea1ab2c",
        "0x82c018f529283075e9d8cdb9c75aca5e1e0064dc335a19a29186aae04a3ea182",
        "0xad9f9ee1bd48c449c9dd7445c79cfb5aad5b9dc678690da848494303d17b8512",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x0001000000000016",
      "value": "0x000000005cdc2123",
      "salt": "0xd5424e727912c610af25c1bf515ea73a388834c5a2f8d91e7db3f1b82af9c5e9",
      "hash": "0x",
      "sorted_hashes": [
        "0x52473c7d1e7dd850dd0c9ae335cee5a63ffefed9106ce6e2cd3eca800421b6af",
        "0x243618b78c5dd6ad0fd8185f771e7020029050867f4496f5a2029649ec8ff4c4",
        "0x19e12e6fbc42a20db4fde852b061eea612e73d89ccdddb525937216f5b609c6d",
        "0x1b52c71d89a7dac886f8550dc04926267ce18cadd1d61c5b94ce8ca19b5dd963",
        "0xab87d05c1b2155acf8d3fbca0ea7eee9a305c872e8e7f8830146a7d12e66f029",
        "0xe598fd9ad7fb03bc8a5c34c9b70e5fa8d11777520a70c8c0b762a7e22ea1ab2c",
        "0x82c018f529283075e9d8cdb9c75aca5e1e0064dc335a19a29186aae04a3ea182",
        "0xad9f9ee1bd48c449c9dd7445c79cfb5aad5b9dc678690da848494303d17b8512",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x0001000000000013",
      "value": "0x48a8509ac72fc0bf508c88a87cf022310d2aba16",
      "salt": "0x957ee5bc82c73d65b56983f6ff7f93189bd2f523828af964adc49db2dadefd94",
      "hash": "0x",
      "sorted_hashes": [
        "0xeb275d9571fd06e81b4a7c3dbd75c2150278907f7200927d2559b7fa2a295a99",
        "0xd52402eed5525b8e22cf4f7c9e4bba9fd39ed72603382975e03d748d98e31339",
        "0x1730dfe3461c213925e63af4ec6287678b5ecbfaee75b7e367bf88e208cf53af",
        "0x1b52c71d89a7dac886f8550dc04926267ce18cadd1d61c5b94ce8ca19b5dd963",
        "0xab87d05c1b2155acf8d3fbca0ea7eee9a305c872e8e7f8830146a7d12e66f029",
        "0xe598fd9ad7fb03bc8a5c34c9b70e5fa8d11777520a70c8c0b762a7e22ea1ab2c",
        "0x82c018f529283075e9d8cdb9c75aca5e1e0064dc335a19a29186aae04a3ea182",
        "0xad9f9ee1bd48c449c9dd7445c79cfb5aad5b9dc678690da848494303d17b8512",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x0001000000000002",
      "value": "0x756e70616964",
      "salt": "0x1b7ec763c8a7cd289cfc6e82ad04f0567c9e798309ee54896b39b26362664b0a",
      "hash": "0x",
      "sorted_hashes": [
        "0xc769fdd8220ed7676e22c2ed278133e0d82069d96fe2972647ad5eef52036364",
        "0xe0ab97a811e20a4f75e0934ed052543a5bbbac177246763d9024e5cf03c3992b",
        "0x3d4fea8bd1a4b9442f8ff729e618d956fcae13b2a05ef6ea4f2aa8e7c763ed2b",
        "0x22cf4b003095411f0f5c2f7795b0ea62c24cbf491f2cd755d102da4287ddbe69",
        "0xf559ea3a8be0f2eedf214c2a90c40377602cbacc67a970fbced9721220944b55",
        "0xe598fd9ad7fb03bc8a5c34c9b70e5fa8d11777520a70c8c0b762a7e22ea1ab2c",
        "0x82c018f529283075e9d8cdb9c75aca5e1e0064dc335a19a29186aae04a3ea182",
        "0xad9f9ee1bd48c449c9dd7445c79cfb5aad5b9dc678690da848494303d17b8512",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x040000000000000a",
      "value": "0x",
      "salt": "0x",
      "hash": "0x121d4f55a467a2258cfcf26ede0644f1eeebca6a3371fae6a5d1b3a52f86188b",
      "sorted_hashes": [
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x030000000000000148a8509ac72fc0bf508c88a87cf022310d2aba16000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c500000004",
      "value": "0xb99c5026e1a44dadd637c9ff9c6083addace4dcc6ac5627d2ec3af2c210676a81dc4b61861e978022579b65a4a381915b4b5db40dc6b56bf52c9d5bf38dcf8f201",
      "salt": "0xe5ce0647e9f2b6ed74f9d6ecf1d2615f2873f214e9880668c6566a34bab8af10",
      "hash": "0x",
      "sorted_hashes": [
        "0xf82a3030cbe661ee9a37c7131b4409ad812e8321d3228ddc986e3eb0d88e2d84",
        "0xd7514d6797c7047114c307cefcede30cd90098eb314b57ec24ee18d8b4126a00",
        "0x875f2ad872d41a59d4989b98e5e3195900ef23efdf539334d95b424c7863c0ea",
        "0x121d4f55a467a2258cfcf26ede0644f1eeebca6a3371fae6a5d1b3a52f86188b"
      ]
    },
    {
      "property": "0x030000000000000148a8509ac72fc0bf508c88a87cf022310d2aba16000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c500000005",
      "value": "0x00",
      "salt": "0x82d4c8ea1eb6b96bce043940fb71e02a95e9cdbcce628f8e80300616a3e4a9c4",
      "hash": "0x",
      "sorted_hashes": [
        "0xca375cfc55d964e57c3933dd27a07fa512f810945553562e1dcb59c2d7c2d789",
        "0x121d4f55a467a2258cfcf26ede0644f1eeebca6a3371fae6a5d1b3a52f86188b"
      ]
    },
    {
      "property": "0x0100000000000004",
      "value": "0x0002a7924d6be77f3521ff20f0bc124d06de86747d0307ff5f3b3c7e90dc9f19",
      "salt": "0xa3d82f5d9a3847f57c92f5c84663be518439b80c319d22f79d37b3c093f68b03",
      "hash": "0x",
      "sorted_hashes": [
        "0x40869a984a9b62b797939b2b61ca186225a6632f8606f38088102e5c7309ad4e",
        "0xff1b5bd5ad51c7d2cca281647b10297d98a7a5a8b959b717954a31dbfb35931b",
        "0xb6146aefb37bf93d1ee8ef5e0279fad21cf5b49b4bc8d408e23c18c7c21b3ccf",
        "0xee9b516a266499d5bc8b8ee5ff1b6b3420a9c3a443faf2380052797821fc0065",
        "0xdbfe7008af5c2392609f197cb22679b1393bfc827510b84f0a6a51f77163483b",
        "0x11be5c687d571fd93f9538e08e216424010d3a96e00954dd85348216dc8d4dbe",
        "0xbf0090247cb2f38603897c97b6969829bdcdf647f4fa3d8b423d26c1145ecf88",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x01000000000000149b62c52a7544be667e7dceef111dcfaf0dd28fcf000000000000000000000000",
      "value": "0xc84a6a6e59be07cecb4d606d52b3937668364a9a48d2df58595c126476c07216",
      "salt": "0xcd21c995fccbe3a12624c8667f74d1703d98db81a0ec0cf44820a6b799fa4f2d",
      "hash": "0x",
      "sorted_hashes": [
        "0xeacddd3d2942f56073343d0ca2967596d9b730be680972a48af2897a14c1bff9",
        "0x8c8782d24d55b474ab4b103f8e5b58d804021991682a48d14625eb89f6f3d1cf",
        "0x9ece58c581584f728ee7a27736591b84260af0889364db4c18b9a1eca833af9a",
        "0x56fb6ec077ba56c0f91b9ad8e0f3f7457d305feacf7fb5671c808a4efccf1dfe",
        "0xb01db608f3a7d02375cc6cb89991cb742cd10c9900d0c1c816b4e9ed1d8462ed",
        "0x11be5c687d571fd93f9538e08e216424010d3a96e00954dd85348216dc8d4dbe",
        "0xbf0090247cb2f38603897c97b6969829bdcdf647f4fa3d8b423d26c1145ecf88",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x01000000000000130000000000000001000000020000000000000000",
      "value": "0xfb3d3e4436f54893b12d67c5438c98ea991c4eeb004610e55bfea0f90ee4fae0",
      "salt": "0x182ea783fb56a91278c6268e3e01b9597ce1c5cb3670c4e970221c5ab60184b9",
      "hash": "0x",
      "sorted_hashes": [
        "0xa34e94b3f51efaa938bd5cd845091f3e4eac19bb1d260887e0c6070b1473f351",
        "0x3cf0e32146a0e91d1beaa2d8418b240f5650b0e17c5f0441c9445c455a38bcd8",
        "0x8cf14d69356d9d1b378b9b6cd918024cec5a2a2340f3df356680c784a9d772ca",
        "0x56fb6ec077ba56c0f91b9ad8e0f3f7457d305feacf7fb5671c808a4efccf1dfe",
        "0xb01db608f3a7d02375cc6cb89991cb742cd10c9900d0c1c816b4e9ed1d8462ed",
        "0x11be5c687d571fd93f9538e08e216424010d3a96e00954dd85348216dc8d4dbe",
        "0xbf0090247cb2f38603897c97b6969829bdcdf647f4fa3d8b423d26c1145ecf88",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x0100000000000013000000000000000100000004",
      "value": "0x0000000000000002",
      "salt": "0x9b114c47790c1de1c03f6c576fc68b8cc39f37ef9b36b90fa77f281b8bd2bbe9",
      "hash": "0x",
      "sorted_hashes": [
        "0xe17dbdda5ab77938350f0ddac81af97f018b90edf87dc4f69b9480fc1ac26a00",
        "0x0e48da0a009000c12abbaa371e4fe65fb94f4a3d0e44a856978cc1942a086f23",
        "0x9ece58c581584f728ee7a27736591b84260af0889364db4c18b9a1eca833af9a",
        "0x56fb6ec077ba56c0f91b9ad8e0f3f7457d305feacf7fb5671c808a4efccf1dfe",
        "0xb01db608f3a7d02375cc6cb89991cb742cd10c9900d0c1c816b4e9ed1d8462ed",
        "0x11be5c687d571fd93f9538e08e216424010d3a96e00954dd85348216dc8d4dbe",
        "0xbf0090247cb2f38603897c97b6969829bdcdf647f4fa3d8b423d26c1145ecf88",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
      ]
    },
    {
      "property": "0x0100000000000001fb3d3e4436f54893b12d67c5438c98ea991c4eeb004610e55bfea0f90ee4fae0000000040000000000000000",
      "value": "0x9b62c52a7544be667e7dceef111dcfaf0dd28fcfc84a6a6e59be07cecb4d606d52b3937668364a9a48d2df58595c126476c07216",
      "salt": "0x0731dc4e67e971ab31287c7f4e775834f5d20d25a9832fb1e6dc330db5fc4af5",
      "hash": "0x",
      "sorted_hashes": [
        "0xf1cb36d36253e6be1b83b8bc872589d8932cbb0ca48fbcdee798cc960cdd46a1",
        "0x6dd60fed54071ba89a42c2e633faa165280566b3125f153557213ba06d5a5072",
        "0xb6146aefb37bf93d1ee8ef5e0279fad21cf5b49b4bc8d408e23c18c7c21b3ccf",
        "0xee9b516a266499d5bc8b8ee5ff1b6b3420a9c3a443faf2380052797821fc0065",
        "0xdbfe7008af5c2392609f197cb22679b1393bfc827510b84f0a6a51f77163483b",
        "0x11be5c687d571fd93f9538e08e216424010d3a96e00954dd85348216dc8d4dbe",
        "0xbf0090247cb2f38603897c97b6969829bdcdf647f4fa3d8b423d26c1145ecf88",
        "0x4764ec359c7fa2393fb3619ffa3489efcbdaa028e2966a4a0dddc2dc63cb7feb"
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
