// +build unit

package documents

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/crypto/pedersen"

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
	ethClient := &ethereum.MockEthClient{}
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

func TestGetDocumentDataProofHash(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	// Arbitrary tree just to pass it to the calculate function
	drTree, err := cd.getSignatureDataTree()
	assert.NoError(t, err)
	docDataRoot, err := cd.CalculateDocumentDataRoot(documenttypes.InvoiceDataTypeUrl, drTree.GetLeaves())
	assert.Nil(t, err)

	cd.GetTestCoreDocWithReset()
	docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, drTree.GetLeaves())
	assert.Nil(t, err)

	signatureTree, err := cd.getSignatureDataTree()
	assert.Nil(t, err)

	valid, err := proofs.ValidateProofSortedHashes(docDataRoot, [][]byte{signatureTree.RootHash()}, docRoot, sha256.New())
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
		Signature:   utils.RandomSlice(65),
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
	// SignerID is not part of the tree
	_, signerLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s].signer_id", SignaturesTreePrefix, signerKey))
	assert.Nil(t, signerLeaf)
	//Leaf contains signature+transitionValidated = 65+1 = 66 bytes
	_, signatureLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signatureLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey), signatureLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, sig.SignatureId...)...), signatureLeaf.Property.CompactName())
	assert.Len(t, signatureLeaf.Value, 66)
	assert.Equal(t, append(sig.Signature, []byte{0}...), signatureLeaf.Value)
}

func TestGetDocumentDocumentDataTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// no data root
	_, _, err = cd.docDataTrees(documenttypes.InvoiceDataTypeUrl, nil)
	assert.Error(t, err)

	// successful tree generation
	dtree, err := cd.getSignatureDataTree()
	assert.NoError(t, err)
	trees, _, err := cd.docDataTrees(documenttypes.InvoiceDataTypeUrl, dtree.GetLeaves())
	assert.Nil(t, err)
	assert.NotNil(t, trees)
	eDataTree := trees[0]
	_, leaf := eDataTree.GetLeafByProperty(SignaturesTreePrefix + ".signatures.length")
	for _, l := range eDataTree.GetLeaves() {
		fmt.Printf("P: %s V: %v", l.Property.ReadableName(), l.Value)
	}
	assert.NotNil(t, leaf)

	_, leaf = eDataTree.GetLeafByProperty(CDTreePrefix + ".current_version")
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
	// Arbitrary tree just to pass it to the calculate function
	drTree, err := cd.getSignatureDataTree()
	assert.NoError(t, err)

	// successful document root generation
	docDataRoot, err := cd.CalculateDocumentDataRoot(documenttypes.InvoiceDataTypeUrl, drTree.GetLeaves())
	assert.NoError(t, err)
	tree, err := cd.DocumentRootTree(documenttypes.InvoiceDataTypeUrl, drTree.GetLeaves())
	assert.NoError(t, err)
	_, leaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, DocumentDataRootField))
	assert.NotNil(t, leaf)
	assert.Equal(t, docDataRoot, leaf.Hash)

	// Get signaturesLeaf
	_, signaturesLeaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField))
	assert.NotNil(t, signaturesLeaf)
	assert.Equal(t, fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField), signaturesLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(DRTreePrefix), CompactProperties(SignaturesRootField)...), signaturesLeaf.Property.CompactName())
}

func TestCoreDocument_GenerateProofs(t *testing.T) {
	h := sha256.New()
	ph := pedersen.NewPedersenHash()
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	testTree, err := cd.DefaultTreeWithPrefix("prefix", []byte{29, 0, 0, 0})
	assert.NoError(t, err)
	props := []proofs.Property{NewLeafProperty("prefix.sample_field", []byte{29, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("prefix.sample_field2", []byte{29, 0, 0, 0, 0, 0, 0, 202})}
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
		Signature:   utils.RandomSlice(65),
	}
	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{sig}
	cdLeaves, err := cd.coredocLeaves(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	zDocDataTree, err := cd.zDocDataTree(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves(), cdLeaves)
	assert.NoError(t, err)
	docRootTree, err := cd.DocumentRootTree(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)

	signerKey := hexutil.Encode(sig.SignatureId)
	signProp := fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey)

	tests := []string{"prefix.sample_field", CDTreePrefix + ".document_identifier", "prefix.sample_field2", CDTreePrefix + ".next_version", signProp}
	pfs, err := cd.CreateZProofs(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves(), tests)
	assert.NoError(t, err)
	assert.Len(t, pfs, 5)

	for idx, test := range tests {
		if strings.Contains(test, SignaturesTreePrefix) {
			signTree, err := cd.getSignatureDataTree()
			assert.NoError(t, err)
			_, l := signTree.GetLeafByProperty(test)
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, pfs[idx].SortedHashes[:1], signTree.RootHash(), h)
			assert.NoError(t, err)
			assert.True(t, valid)
		} else {
			assert.Len(t, pfs[idx].Hashes, 7)
			_, l := zDocDataTree.GetLeafByProperty(test)
			// Validate from leaf to zDocDataRoot as it uses pedersen hash
			valid, err := proofs.ValidateProofHashes(l.Hash, pfs[idx].Hashes[:len(pfs[idx].Hashes)-2], zDocDataTree.RootHash(), ph)
			assert.NoError(t, err)
			assert.True(t, valid)
			// Validate from zDocDataRoot to docRoot as it uses sha hash
			valid, err = proofs.ValidateProofHashes(zDocDataTree.RootHash(), pfs[idx].Hashes[len(pfs[idx].Hashes)-2:], docRootTree.RootHash(), h)
			assert.NoError(t, err)
			assert.True(t, valid)
		}
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
	_, err = cd.DeleteAttribute(key, true, nil)
	assert.Error(t, err)

	// success
	attr, err := NewAttribute(label, AttrString, value)
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
	attr, err = NewAttribute(label, AttrDecimal, nvalue)
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
    "document_id": "0xeb8e023d7e68f012d0d20265b95a4fd6965c0a225eec08a71cfd449def7ed384",
    "version_id": "0x1f5d8246e58f56326caa4d57a0d9e73ddaa286b17dde0588562136ec05375a4f",
		"document_root": "0xf15441df88f11a38e01806579b0c10fe90db7e0c8c6e78e90606a569cc967e3d"
  },
  "field_proofs": [
    {
      "property": "0x000100000000000e",
      "value": "0x0006aaf7c8516d0c000000000000000000000000000000000000000000000000",
      "salt": "0x03ad6dc0291471f29cf439acd446090b74433d19dd0ba1f3d8597a468031659b",
      "hash": "0x",
      "sorted_hashes": [
        "0xe045afed673da7262e803b39c27feedf09824c9c8c1f8858694b8da32c338c84",
        "0x7eb570b329e6511f00222727eefb1ff30aa5171fa4889ae2a969fad1d686ebef",
        "0x838822fcd4d58e06e8ffe61d41379d2618c5b2f7cf42bdd98e9f943e8717e79c",
        "0x1c19e26fe82343282eab0331aa48ef4429113471cf775cea873906f72abe2c09",
        "0x18b16a9b0e833f10f998a72a1bb50e2726fe29856714c36c81c8f203d57c98b9",
        "0xac4abb4b3b8df91e9551a047ea5223e18beee6c4a8efcad45cca6fa4c2693306",
        "0xb6bf7839a75590bee871573b4fad0bb688f5c6a29c992214272c82199e7d86dd",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x000100000000000d",
      "value": "0x455552",
      "salt": "0x4caa54d2dd4a49458a9f4647dca1b6605028b90629b1e60d15109e3c1eb0ffa6",
      "hash": "0x",
      "sorted_hashes": [
        "0xafff7542a9c76020daf37e528260b8e57f17f04621952792c59aefe5ffde5b09",
        "0x7eb570b329e6511f00222727eefb1ff30aa5171fa4889ae2a969fad1d686ebef",
        "0x838822fcd4d58e06e8ffe61d41379d2618c5b2f7cf42bdd98e9f943e8717e79c",
        "0x1c19e26fe82343282eab0331aa48ef4429113471cf775cea873906f72abe2c09",
        "0x18b16a9b0e833f10f998a72a1bb50e2726fe29856714c36c81c8f203d57c98b9",
        "0xac4abb4b3b8df91e9551a047ea5223e18beee6c4a8efcad45cca6fa4c2693306",
        "0xb6bf7839a75590bee871573b4fad0bb688f5c6a29c992214272c82199e7d86dd",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x0001000000000016",
      "value": "0x000000005cffba11",
      "salt": "0xaabd761207469b15cdeaa8bf41a0db621a2d50dde83d2b563d15727dd3e595f9",
      "hash": "0x",
      "sorted_hashes": [
        "0x9f859096730418e38bfdf1c610e27f42a015e323dd33085c9b6da903a9c5d274",
        "0x8634559cc0738a62f81443624d7134fc1b3c8cb5395cf68b7922057632568d9e",
        "0x6919b0a98f5b657708af9a6145198efc1abebcb3522b9504fad31b21c4d108ee",
        "0xbf4178a54a6ff56cb31b6c96e0e4e76b09936f79d91ec37c989b0d2ff4bba00b",
        "0xf77331d7cf79d218150f72490579b6c16a65f3f2c14a86b7b1913d321b16497b",
        "0xac4abb4b3b8df91e9551a047ea5223e18beee6c4a8efcad45cca6fa4c2693306",
        "0xb6bf7839a75590bee871573b4fad0bb688f5c6a29c992214272c82199e7d86dd",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x0001000000000013",
      "value": "0x359679e690c685f88ad16265deb28001f43c0ac5",
      "salt": "0x1c7c7dd52966804acfbef10707ceb06455b609ca566cf039e4a64d6cdd3d9145",
      "hash": "0x",
      "sorted_hashes": [
        "0x8f081964af6741a547ac75a246604e96d0887acb4789e3863bd09331729f30e3",
        "0x0d5d44a8a511e0bc27527131236a1da4373bac40717c11660f078ed26672d44d",
        "0xb0911d74d8ea9bc7c770c039621121cf6511993df401f61db4e93673fdec87d2",
        "0xbf4178a54a6ff56cb31b6c96e0e4e76b09936f79d91ec37c989b0d2ff4bba00b",
        "0xf77331d7cf79d218150f72490579b6c16a65f3f2c14a86b7b1913d321b16497b",
        "0xac4abb4b3b8df91e9551a047ea5223e18beee6c4a8efcad45cca6fa4c2693306",
        "0xb6bf7839a75590bee871573b4fad0bb688f5c6a29c992214272c82199e7d86dd",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x0001000000000002",
      "value": "0x756e706169640000000000000000000000000000000000000000000000000000",
      "salt": "0xe0cb986cff7c394a3027a33c526450b81a39df6fe9493fff490256b17ff03ea5",
      "hash": "0x",
      "sorted_hashes": [
        "0x05ff8535ec7c547b7bc7df19f3b10d74a7adfa9a382ed92c8fc3211b6b89e83d",
        "0xe826b2883ff3643572c815f51dfcc710b3ce8bb8fa5ab5dc32188e84c7894906",
        "0xe4507eba11dfcd79c7e85ce73254e0d7f4865dfd3b708e36e698487a21af72fe",
        "0x2eca620cd4c13bc3e545dec20dfd57465a1ce59cfecdb53e202e9b5442eb4fba",
        "0x18b16a9b0e833f10f998a72a1bb50e2726fe29856714c36c81c8f203d57c98b9",
        "0xac4abb4b3b8df91e9551a047ea5223e18beee6c4a8efcad45cca6fa4c2693306",
        "0xb6bf7839a75590bee871573b4fad0bb688f5c6a29c992214272c82199e7d86dd",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x040000000000000a",
      "value": "0x",
      "salt": "0x",
      "hash": "0x8191e5e563ed8b25dd46fd383df6a933cdebd6cbcf0f79f8f70e33510adceb88",
      "sorted_hashes": [
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x0300000000000001359679e690c685f88ad16265deb28001f43c0ac5000000000000000000000000746c4d8464ad40caadc76c2c0b31393c6ae0d6c5",
      "value": "0x44acda5d9f2c27549b168859be8cb0cf5a100931a52b2fc54c50ee080a5a7d590068200a00bead4c119ade7d0dd2ab6382138d8e0a5566cfbd285cc6439ea5100000",
      "salt": "0x71fd192a8a0853e6988e9633fca89dacdae87fbd1051f4537c98922a8e86c356",
      "hash": "0x",
      "sorted_hashes": [
        "0x425bd50a1f4165b03b001a3cf14588d5f44e1f5cdc205679edc0788034355ea7",
        "0x8191e5e563ed8b25dd46fd383df6a933cdebd6cbcf0f79f8f70e33510adceb88"
      ]
    },
    {
      "property": "0x0100000000000004",
      "value": "0x6033db92acfd319cbc57660b8f40874033976a188b311542f8468be3be086760",
      "salt": "0x7c4d38cc36790588c2435e1d8fc6945caa1abdedcc65c48997be5a317fe31e69",
      "hash": "0x",
      "sorted_hashes": [
        "0xe66d7cfb613d1815f192af9f0df3b1792b0c2e09548d72b7fe5880e8a66997e8",
        "0xf9f0707c87cea783a63d78349578eae81548d2e832b4d33fe4dd904a506c51cb",
        "0x9789dcad2d98b708ea28ee3419e258281aafbdef0352c24b57775a6dc59559b9",
        "0x5c54b8fb4258a6afaefbb0fdb92a4b22da30574eb0f772bab0398c8a7a3529e4",
        "0xc0da6473d496864bb719539dd216b412f62a8272155bfddb53635a472fd20857",
        "0xce5ec833bb5cddc439a22b9f6922ec230eae6d2d3dc0c9ac2b81237523ba0b36",
        "0x02d0f508c95d861e998344585d9d44c36c993afc43376d1b38123ad7a28f0f0b",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x01000000000000144e498bd371833d6b910513a6e15f1bd8e97fbeda000000000000000000000000",
      "value": "0x67d7734bed88b7555fdd2e2121c80ed57bac7951fc6b0a322a3014079b668a4d",
      "salt": "0xffbc8554e0f239329b67edff515b069c95cf6ae55aa435b08b47c1d8352f1f35",
      "hash": "0x",
      "sorted_hashes": [
        "0x527128e18d942c764096d605906650ecb4aa7273aa7cbcbc5b4281fbf05d9fa0",
        "0x264c2e7829da8196a7ec00529ed6f9124079dbffe41184824f30a091fa56d097",
        "0x6582cd7861bb428e920a2c82a2919c5a1307a5e7f746e6e5dbfdfd73ebc7d519",
        "0x7f4b9794449efd3f078b91baad67cb1d8467b061e9605a0481fa83ef23a1bbc0",
        "0xd420d40ac599ff816fa8f1f1b50be6843224b36e6f12ca5a76ef5d0080328263",
        "0xce5ec833bb5cddc439a22b9f6922ec230eae6d2d3dc0c9ac2b81237523ba0b36",
        "0x02d0f508c95d861e998344585d9d44c36c993afc43376d1b38123ad7a28f0f0b",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x01000000000000130000000000000001000000020000000000000000",
      "value": "0x8a963e5d9dda1a9ab943ac92d528a229e06393e3f619de0b45a0a2f76f749ffc",
      "salt": "0x782ad260b8695c598a546ab756afc127312b8028aa0c8e0e970cfc954179f5a4",
      "hash": "0x",
      "sorted_hashes": [
        "0x101a8b4e71bbf529b9081129bab31055942d1d8f028545304688cfaf3c3c3033",
        "0x67cb8534b9856c945d762ff432bb62e32413bdd66b4a29dfc8f9b6d6675f686c",
        "0xe3474af96936523e6e1f79fd1344baef83ff1eaa2f987ff15f2e121dc8ae2760",
        "0x6396647dc7be668b703aadaab64847f3f3c07405dcdd0a9ddac90637254f3f03",
        "0xd420d40ac599ff816fa8f1f1b50be6843224b36e6f12ca5a76ef5d0080328263",
        "0xce5ec833bb5cddc439a22b9f6922ec230eae6d2d3dc0c9ac2b81237523ba0b36",
        "0x02d0f508c95d861e998344585d9d44c36c993afc43376d1b38123ad7a28f0f0b",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x0100000000000013000000000000000100000004",
      "value": "0x0000000000000002",
      "salt": "0x460289923e56900a0fe1632e660d48728b489fcd8ed4d36bc0465a01213858b5",
      "hash": "0x",
      "sorted_hashes": [
        "0x8c50e4247a724a840d1942195657b520c216259e32b7691a28cf89bc90c1cdf4",
        "0xbcbaf915a261331264d489daedd86199914a6feed11c81389bee3aa055dc4061",
        "0xe3474af96936523e6e1f79fd1344baef83ff1eaa2f987ff15f2e121dc8ae2760",
        "0x6396647dc7be668b703aadaab64847f3f3c07405dcdd0a9ddac90637254f3f03",
        "0xd420d40ac599ff816fa8f1f1b50be6843224b36e6f12ca5a76ef5d0080328263",
        "0xce5ec833bb5cddc439a22b9f6922ec230eae6d2d3dc0c9ac2b81237523ba0b36",
        "0x02d0f508c95d861e998344585d9d44c36c993afc43376d1b38123ad7a28f0f0b",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
      ]
    },
    {
      "property": "0x01000000000000018a963e5d9dda1a9ab943ac92d528a229e06393e3f619de0b45a0a2f76f749ffc000000040000000000000000",
      "value": "0x4e498bd371833d6b910513a6e15f1bd8e97fbeda67d7734bed88b7555fdd2e2121c80ed57bac7951fc6b0a322a3014079b668a4d",
      "salt": "0x786093ef34d80ff3d1a68c68f8d2cb90b928925d59c49a90f0e7ef7582e4f1c3",
      "hash": "0x",
      "sorted_hashes": [
        "0x3073b29bc0b8dd7f189a29318c38f55ea3ed9958b577417dcaeb76e0c5eae74a",
        "0x33e89ba2b64cd55fa20002879289110c6b78e7dfdbfd22c7c011e793707ea631",
        "0x5c5601997e2eb798dd827d6ee12992b52b3d283f46f08e9a531e94d30ea29e06",
        "0x5c54b8fb4258a6afaefbb0fdb92a4b22da30574eb0f772bab0398c8a7a3529e4",
        "0xc0da6473d496864bb719539dd216b412f62a8272155bfddb53635a472fd20857",
        "0xce5ec833bb5cddc439a22b9f6922ec230eae6d2d3dc0c9ac2b81237523ba0b36",
        "0x02d0f508c95d861e998344585d9d44c36c993afc43376d1b38123ad7a28f0f0b",
        "0xf6609fda813892406e3b8a676caa2a3b633edaee6fa0ee2fc2591c4280811a56"
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
		assert.True(t, valid, fmt.Sprintf("Failed for proof %d", i))
	}

}
