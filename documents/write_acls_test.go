//go:build unit
// +build unit

package documents

import (
	"bytes"
	"crypto/sha256"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	genericpb "github.com/centrifuge/centrifuge-protobufs/gen/go/generic"
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestWriteACLs_getChangedFields_different_types(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	ocd := cd.Document
	ncd := genericpb.GenericData{
		Scheme: []byte("generic"),
	}

	oldTree := getTree(t, &ocd, "", nil)
	newTree := getTree(t, &ncd, "", nil)

	cf := GetChangedFields(oldTree, newTree)
	// cf length should be len(ocd) and len(ncd) = 10 changed field
	assert.Len(t, cf, 10)
}

func TestWriteACLs_getChangedFields_same_document(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	ocd := cd.Document
	oldTree := getTree(t, &ocd, "", nil)
	newTree := getTree(t, &ocd, "", nil)
	cf := GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 0)

	// check author field
	ocd.Author = utils.RandomSlice(20)
	oldTree = getTree(t, &ocd, "", nil)
	newTree = getTree(t, &ocd, "", nil)
	cf = GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 0)
}

func testExpectedProps(t *testing.T, cf []ChangedField, eprops map[string]struct{}) {
	for _, f := range cf {
		_, ok := eprops[hexutil.Encode(f.Property)]
		if !ok {
			assert.Failf(t, "", "expected %x property to be present", f.Property)
		}
	}
}

func TestWriteACLs_getChangedFields_with_core_document(t *testing.T) {
	doc, err := newCoreDocument()
	assert.NoError(t, err)
	ndoc, err := doc.PrepareNewVersion([]byte("po"), CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{testingidentity.GenerateRandomDID()}}, nil)
	assert.NoError(t, err)

	// preparing new version would have changed the following properties

	// current_version
	// previous_version
	// next_version
	// previous_root
	// current pre image
	// next pre image

	// read_rules.roles
	// read_rules.action
	// transition_rules.RuleKey
	// (transition_rules.Roles
	// transition_rules.MatchType
	// transition_rules.Action
	// transition_rules.Field
	// transition_rules.ComputeFields
	// transition_rules.ComputeTargetField
	// transition_rules.ComputeCode) x 2
	// roles + 2
	oldTree := getTree(t, &doc.Document, "", nil)
	newTree := getTree(t, &ndoc.Document, "", nil)
	cf := GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 23)
	rprop := append(ndoc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop2 := append(ndoc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	eprops := map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 4}):  {},
		hexutil.Encode([]byte{0, 0, 0, 3}):  {},
		hexutil.Encode([]byte{0, 0, 0, 16}): {},
		hexutil.Encode([]byte{0, 0, 0, 2}):  {},
		hexutil.Encode([]byte{0, 0, 0, 22}): {},
		hexutil.Encode([]byte{0, 0, 0, 23}): {},
		hexutil.Encode([]byte{0, 0, 0, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 3}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 5}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 6}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 7}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 8}):                         {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop...)):                                            {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop2...)):                                           {},
	}

	testExpectedProps(t, cf, eprops)

	// prepare new version with out new collaborators
	// this should only change
	// current_version
	// previous_version
	// next_version
	// previous_root
	// current pre image
	// next pre image
	doc = ndoc
	ndoc, err = doc.PrepareNewVersion([]byte("po"), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document, "", nil)
	newTree = getTree(t, &ndoc.Document, "", nil)
	cf = GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 5)
	eprops = map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 4}):  {},
		hexutil.Encode([]byte{0, 0, 0, 3}):  {},
		hexutil.Encode([]byte{0, 0, 0, 16}): {},
		hexutil.Encode([]byte{0, 0, 0, 2}):  {},
		hexutil.Encode([]byte{0, 0, 0, 22}): {},
		hexutil.Encode([]byte{0, 0, 0, 23}): {},
	}
	testExpectedProps(t, cf, eprops)

	// test with different document
	// this will change
	// document identifier
	// current version
	// previous version
	// next version
	// previous_root
	// current pre image
	// next pre image
	// roles (new doc will have empty role while old one has two roles)
	// read_rules (new doc will have empty read_rules while old one has read_rules)
	// transition_rules (new doc will have empty transition_rules while old one has 2 transition_rules)
	doc = ndoc
	ndoc, err = newCoreDocument()
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document, "", nil)
	newTree = getTree(t, &ndoc.Document, "", nil)
	cf = GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 24)
	rprop = append(doc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop2 = append(doc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	eprops = map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 5}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 3}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 6}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 7}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 8}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 9}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 4}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 3}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 16}):                                                             {},
		hexutil.Encode([]byte{0, 0, 0, 2}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 22}):                                                             {},
		hexutil.Encode([]byte{0, 0, 0, 23}):                                                             {},
		hexutil.Encode([]byte{0, 0, 0, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}):                         {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop...)):                                            {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop2...)):                                           {},
	}
	testExpectedProps(t, cf, eprops)

	// add different roles and read rules and check
	// this will change
	// current version
	// previous version
	// next version
	// previous_root
	// current pre image
	// next pre image
	// roles (new doc will have 2 new roles different from 2 old roles)
	// read_rules
	// transition_rules
	ndoc, err = ndoc.PrepareNewVersion([]byte("po"), CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{testingidentity.GenerateRandomDID()}}, nil)
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document, "", nil)
	newTree = getTree(t, &ndoc.Document, "", nil)
	cf = GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 15)
	rprop = append(doc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop2 = append(doc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop3 := append(ndoc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop4 := append(ndoc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	eprops = map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 9}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 4}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 3}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 16}):                                                             {},
		hexutil.Encode([]byte{0, 0, 0, 2}):                                                              {},
		hexutil.Encode([]byte{0, 0, 0, 22}):                                                             {},
		hexutil.Encode([]byte{0, 0, 0, 23}):                                                             {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop...)):                                            {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop2...)):                                           {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop3...)):                                           {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop4...)):                                           {},
	}
	testExpectedProps(t, cf, eprops)
}

func TestWriteACLs_getChangedFields_document(t *testing.T) {
	// no change
	doc := &genericpb.GenericData{
		Scheme: []byte("Generic"),
	}

	oldTree := getTree(t, doc, "", nil)
	newTree := getTree(t, doc, "", nil)
	cf := GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 0)

	// updated doc
	ndoc := &genericpb.GenericData{
		Scheme: []byte("generic"),
	}

	oldTree = getTree(t, doc, "", nil)
	newTree = getTree(t, ndoc, "", nil)
	cf = GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 1)
	eprops := map[string]ChangedField{
		hexutil.Encode([]byte{0, 0, 0, 1}): {
			Property: []byte{0, 0, 0, 1},
			Name:     "scheme",
			Old:      []byte{71, 101, 110, 101, 114, 105, 99},
			New:      []byte{103, 101, 110, 101, 114, 105, 99},
		},
	}

	for _, f := range cf {
		ef, ok := eprops[hexutil.Encode(f.Property)]
		if !ok {
			t.Fatalf("expected %x property change", f.Property)
		}

		assert.True(t, reflect.DeepEqual(f, ef))
	}

	// completely new doc
	// this should give 5 property changes
	ndoc = new(genericpb.GenericData)
	oldTree = getTree(t, doc, "", nil)
	newTree = getTree(t, ndoc, "", nil)
	cf = GetChangedFields(oldTree, newTree)
	assert.Len(t, cf, 1)
	eprps := map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 1}): {},
	}
	testExpectedProps(t, cf, eprps)
}

func getTree(t *testing.T, doc proto.Message, prefix string, compact []byte) *proofs.DocumentTree {
	var prop proofs.Property
	if prefix != "" {
		prop = proofs.Property{
			Text:    prefix,
			Compact: compact,
		}
	}
	tr, err := proofs.NewDocumentTree(proofs.TreeOptions{
		ParentPrefix:      prop,
		CompactProperties: true,
		EnableHashSorting: true,
		Hash:              sha256.New(),
		Salts: func(compact []byte) (bytes []byte, e error) {
			return utils.RandomSlice(32), nil
		},
	})
	assert.NoError(t, err)
	tree := &tr
	assert.NoError(t, tree.AddLeavesFromDocument(doc))
	assert.NoError(t, tree.Generate())
	return tree
}

func TestCoreDocument_transitionRuleForAccount(t *testing.T) {
	doc, err := newCoreDocument()
	assert.NoError(t, err)
	id := testingidentity.GenerateRandomDID()
	rules := doc.TransitionRulesFor(id)
	assert.Len(t, rules, 0)

	// add roles and rules
	_, rule := createTransitionRules(t, doc, id, nil, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX)
	rules = doc.TransitionRulesFor(id)
	assert.Len(t, rules, 1)
	assert.Equal(t, *rule, rules[0])

	// wrong id
	rules = doc.TransitionRulesFor(testingidentity.GenerateRandomDID())
	assert.Len(t, rules, 0)
}

func createTransitionRules(_ *testing.T, doc *CoreDocument, id identity.DID, field []byte, matchType coredocumentpb.FieldMatchType) (*coredocumentpb.Role, *coredocumentpb.TransitionRule) {
	role := newRoleWithRandomKey()
	role.Collaborators = append(role.Collaborators, id[:])
	rule := &coredocumentpb.TransitionRule{
		RuleKey:   utils.RandomSlice(32),
		Roles:     [][]byte{role.RoleKey},
		Field:     field,
		MatchType: matchType,
		Action:    coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
	}
	doc.Document.TransitionRules = append(doc.Document.TransitionRules, rule)
	doc.Document.Roles = append(doc.Document.Roles, role)
	return role, rule
}

func prepareDocument(t *testing.T) (*CoreDocument, identity.DID, identity.DID, string) {
	doc, err := newCoreDocument()
	assert.NoError(t, err)
	docType := documenttypes.GenericDataTypeUrl
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()

	// id1 will have rights to update all the fields in the core document
	createTransitionRules(t, doc, id1, CompactProperties(CDTreePrefix), coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX)

	// id2 will have write access to only identifiers
	// id2 is the bad actor
	fields := [][]byte{
		{0, 0, 0, 4},
		{0, 0, 0, 3},
		{0, 0, 0, 16},
		{0, 0, 0, 2},
		{0, 0, 0, 22},
		{0, 0, 0, 23},
	}

	for _, f := range fields {
		createTransitionRules(t, doc, id2, append(CompactProperties(CDTreePrefix), f...), coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT)
	}

	return doc, id1, id2, docType
}

func TestWriteACLs_validateTransitions_roles_read_rules(t *testing.T) {
	doc, id1, id2, docType := prepareDocument(t)

	// prepare a new version of the document with out collaborators
	ndoc, err := doc.PrepareNewVersion([]byte("generic"), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	// if this was changed by the id1, everything should be fine
	assert.NoError(t, doc.CollaboratorCanUpdate(ndoc, id1, docType))

	// if this was changed by id2, it should still be okay since roles would not have changed
	assert.NoError(t, doc.CollaboratorCanUpdate(ndoc, id2, docType))

	// prepare the new document with a new collaborator, this will trigger read_rules and roles update
	ndoc, err = doc.PrepareNewVersion([]byte("generic"), CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{testingidentity.GenerateRandomDID()}}, nil)
	assert.NoError(t, err)

	// should not error out if the change was done by id1
	assert.NoError(t, doc.CollaboratorCanUpdate(ndoc, id1, docType))

	// this should fail since id2 has no write permission to roles, read_rules, and transition rules
	err = doc.CollaboratorCanUpdate(ndoc, id2, docType)
	assert.Error(t, err)
	// we should have 3 errors
	// 1. update to roles
	// 2. update to read_rules
	// 3. update to read_rules action
	assert.Equal(t, 18, errors.Len(err))

	// check with some random collaborator who has no permission at all
	err = doc.CollaboratorCanUpdate(ndoc, testingidentity.GenerateRandomDID(), docType)
	assert.Error(t, err)
	// error should all have field changes
	// all the identifier changes = 6
	// role changes = 2
	// read_rule changes = 2
	// transition rule changes = 10
	// total = 9
	assert.Equal(t, 23, errors.Len(err))
}

func TestWriteACLs_validate_transitions_nfts(t *testing.T) {
	doc, id1, id2, docType := prepareDocument(t)

	// update nfts alone check for validation
	// this should only change nfts
	registry := testingidentity.GenerateRandomDID()
	ndoc, err := doc.AddNFT(false, registry.ToAddress(), utils.RandomSlice(32), true)
	assert.NoError(t, err)

	// if id1 changed it, it should be okay
	assert.NoError(t, doc.CollaboratorCanUpdate(ndoc, id1, docType))

	// if id2  made the change, it should error out with one invalid transition
	err = doc.CollaboratorCanUpdate(ndoc, id2, docType)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))

	// add a specific rule that allow id2 to update specific nft registry
	field := append(registry.ToAddress().Bytes(), make([]byte, 12)...)
	field = append(CompactProperties(CDTreePrefix), append([]byte{0, 0, 0, 20}, field...)...)
	createTransitionRules(t, doc, id2, field, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT)
	ndoc, err = doc.AddNFT(false, registry.ToAddress(), utils.RandomSlice(32), true)
	assert.NoError(t, err)

	// if id1 changed it, it should be okay
	assert.NoError(t, doc.CollaboratorCanUpdate(ndoc, id1, docType))

	// if id2 should be okay since we added a specific registry
	assert.NoError(t, doc.CollaboratorCanUpdate(ndoc, id2, docType))

	// id2 went rogue and updated nft for different registry
	registry2 := testingidentity.GenerateRandomDID()
	ndoc1, err := ndoc.AddNFT(false, registry2.ToAddress(), utils.RandomSlice(32), true)
	assert.NoError(t, err)

	// if id1 changed it, it should be okay
	assert.NoError(t, ndoc.CollaboratorCanUpdate(ndoc1, id1, docType))

	// if id2 is allowed to change only nft with specific registry
	// this should trigger 1 error
	err = ndoc.CollaboratorCanUpdate(ndoc1, id2, docType)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))

	// add a rule for id2 that will allow any nft update
	field = append(CompactProperties(CDTreePrefix), []byte{0, 0, 0, 20}...)
	createTransitionRules(t, ndoc1, id2, field, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX)

	ndoc2, err := ndoc1.AddNFT(false, testingidentity.GenerateRandomDID().ToAddress(), utils.RandomSlice(32), true)
	assert.NoError(t, err)

	// id1 change should be fine
	assert.NoError(t, ndoc1.CollaboratorCanUpdate(ndoc2, id1, docType))

	// id2 change should be fine since id2 has a rule allowing nft update
	assert.NoError(t, ndoc1.CollaboratorCanUpdate(ndoc2, id2, docType))

	// now make a change that will trigger read rules and roles as well
	ndoc2, err = ndoc1.AddNFT(true, testingidentity.GenerateRandomDID().ToAddress(), utils.RandomSlice(32), true)
	assert.NoError(t, err)

	// id1 change should be fine
	assert.NoError(t, ndoc1.CollaboratorCanUpdate(ndoc2, id1, docType))

	// id2 change will be invalid since with grant access, roles and read_rules will be updated
	// this will lead to 3 errors
	// 1. roles
	// 2. read_rules.roles
	// 3. read_rules.action
	err = ndoc1.CollaboratorCanUpdate(ndoc2, id2, docType)
	assert.Error(t, err)
	assert.Equal(t, 3, errors.Len(err))
}

func testDocumentChange(t *testing.T, cd *CoreDocument, id identity.DID, doc1, doc2 proto.Message, prefix string, compact []byte) error {
	oldTree := getTree(t, doc1, prefix, compact)
	newTree := getTree(t, doc2, prefix, compact)

	cf := GetChangedFields(oldTree, newTree)
	rules := cd.TransitionRulesFor(id)
	return ValidateTransitions(rules, cf)
}

func TestWriteACLs_validTransitions_document_data(t *testing.T) {
	doc, id1, id2, _ := prepareDocument(t)
	g := genericpb.GenericData{
		Scheme: []byte("Generic"),
	}

	prefix, compact := "generic", []byte{0, 1, 0, 0}
	// add rules to id1 to update anything on the generic document
	createTransitionRules(t, doc, id1, compact, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX)

	// id2 can only update comment on document and nothing else
	createTransitionRules(t, doc, id2, append(compact, []byte{0, 0, 0, 52}...), coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT)

	g2 := g
	g2.Scheme = []byte("generic")

	// check if id1 made the update
	assert.NoError(t, testDocumentChange(t, doc, id1, &g, &g2, prefix, compact))

	// id2 should fail since it can only change comment
	// errors should be 1
	err := testDocumentChange(t, doc, id2, &g, &g2, prefix, compact)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))
}

func TestWriteACLs_initTransitionRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.initTransitionRules(nil, nil)
	assert.Nil(t, cd.Document.Roles)
	assert.Nil(t, cd.Document.TransitionRules)

	collab := []identity.DID{testingidentity.GenerateRandomDID()}
	cd.initTransitionRules(nil, collab)
	assert.Len(t, cd.Document.TransitionRules, 2)
	assert.Len(t, cd.Document.Roles, 1)

	cd.initTransitionRules(nil, collab)
	assert.Len(t, cd.Document.TransitionRules, 2)
	assert.Len(t, cd.Document.Roles, 1)
}

func roleExistsInRules(t *testing.T, cd *CoreDocument, role []byte, checkRoleCount bool, roleCount int) {
	fieldMap := defaultRuleFieldProps()
	for _, rule := range cd.Document.TransitionRules {
		assert.False(t, deleteFieldIfRoleExists(rule, role, fieldMap))
		if checkRoleCount {
			assert.Len(t, rule.Roles, roleCount)
		}
	}
}

func Test_addDefaultRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.Len(t, cd.Document.TransitionRules, 0)

	// no rules
	role := utils.RandomSlice(32)
	cd.addDefaultRules(role)
	assert.Len(t, cd.Document.TransitionRules, 7)
	roleExistsInRules(t, cd, role, true, 1)

	// all rules present
	role2 := utils.RandomSlice(32)
	cd.addDefaultRules(role2)
	assert.Len(t, cd.Document.TransitionRules, 7)
	roleExistsInRules(t, cd, role2, true, 2)

	// some rules present
	cd.Document.TransitionRules[0].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX
	cd.Document.TransitionRules[1].Field = utils.RandomSlice(5)
	cd.addDefaultRules(role)
	assert.Len(t, cd.Document.TransitionRules, 9)
	roleExistsInRules(t, cd, role, false, 1)
	cd.addDefaultRules(role2)
	roleExistsInRules(t, cd, role2, false, 1)
}

func TestCoreDocument_AddTransitionRuleForAttribute(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	attrKey, err := AttrKeyFromLabel("test1")
	assert.NoError(t, err)
	roleKey := utils.RandomSlice(32)

	// no role exists
	r, err := cd.AddTransitionRuleForAttribute(roleKey, attrKey)
	assert.EqualError(t, err, ErrRoleNotExist.Error())
	assert.Nil(t, r)
	_, err = cd.GetTransitionRule(utils.RandomSlice(32))
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrTransitionRuleMissing, err))
	assert.Len(t, cd.Document.TransitionRules, 0)

	// create new set of rules
	_, err = cd.AddRole(hexutil.Encode(roleKey), []identity.DID{testingidentity.GenerateRandomDID()})
	assert.NoError(t, err)
	r, err = cd.AddTransitionRuleForAttribute(roleKey, attrKey)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Len(t, cd.Document.TransitionRules, 8) // 7(default rules) + 1(attribute rule) = 8
	roleExistsInRules(t, cd, roleKey, true, 1)
	gr, err := cd.GetTransitionRule(r.RuleKey)
	assert.NoError(t, err)
	assert.Equal(t, r, gr)

	// add another attr with same role
	attrKey1, err := AttrKeyFromLabel("test2")
	assert.NoError(t, err)
	r, err = cd.AddTransitionRuleForAttribute(roleKey, attrKey1)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Len(t, cd.Document.TransitionRules, 9) // 8(old rules) + 1(new rule) = 9
	roleExistsInRules(t, cd, roleKey, true, 1)
	gr, err = cd.GetTransitionRule(r.RuleKey)
	assert.NoError(t, err)
	assert.Equal(t, r, gr)

	// new role with old attr
	roleKey = utils.RandomSlice(32)
	_, err = cd.AddRole(hexutil.Encode(roleKey), []identity.DID{testingidentity.GenerateRandomDID()})
	assert.NoError(t, err)
	r, err = cd.AddTransitionRuleForAttribute(roleKey, attrKey)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Len(t, cd.Document.TransitionRules, 10) // 9(old rules) + 1(new rule) = 10
	roleExistsInRules(t, cd, roleKey, false, 1)
	gr, err = cd.GetTransitionRule(r.RuleKey)
	assert.NoError(t, err)
	assert.Equal(t, r, gr)
}

func setupRules(t *testing.T) (*CoreDocument, *coredocumentpb.TransitionRule, []byte) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	attrKey, err := AttrKeyFromLabel("test1")
	assert.NoError(t, err)
	roleKey := utils.RandomSlice(32)
	_, err = cd.AddRole(hexutil.Encode(roleKey), []identity.DID{testingidentity.GenerateRandomDID()})
	assert.NoError(t, err)
	rule, err := cd.AddTransitionRuleForAttribute(roleKey, attrKey)
	assert.NoError(t, err)
	assert.True(t, byteutils.ContainsBytesInSlice(rule.Roles, roleKey))
	assert.True(t, isRoleAssignedToRules(cd, roleKey))
	return cd, rule, roleKey
}

func Test_isRoleReUsed(t *testing.T) {
	cd, rule, roleKey := setupRules(t)

	// delete rule
	assert.NotNil(t, cd.deleteRule(rule.RuleKey))
	assert.False(t, isRoleAssignedToRules(cd, roleKey))
}

func roleNotExists(cd *CoreDocument, roleID []byte) bool {
	for _, rule := range cd.Document.TransitionRules {
		if byteutils.ContainsBytesInSlice(rule.Roles, roleID) {
			return false
		}
	}

	return true
}

func Test_deleteRoleFromDefaultRules(t *testing.T) {
	cd, rule, roleID := setupRules(t)

	assert.Len(t, cd.Document.TransitionRules, 8)
	assert.Equal(t, cd.Document.TransitionRules[7], rule)
	assert.False(t, roleNotExists(cd, roleID))
	cd.deleteRoleFromDefaultRules(roleID)
	assert.NotNil(t, cd.deleteRule(rule.RuleKey))
	assert.Len(t, cd.Document.TransitionRules, 7)
	assert.True(t, roleNotExists(cd, roleID))
}

func Test_deleteRule(t *testing.T) {
	cd, _, _ := setupRules(t)

	// delete first rule
	assert.Len(t, cd.Document.TransitionRules, 8)
	assert.NotNil(t, cd.deleteRule(cd.Document.TransitionRules[0].RuleKey))
	assert.Len(t, cd.Document.TransitionRules, 7)

	// delete rule in between
	assert.NotNil(t, cd.deleteRule(cd.Document.TransitionRules[3].RuleKey))
	assert.Len(t, cd.Document.TransitionRules, 6)

	// delete last rule
	assert.NotNil(t, cd.deleteRule(cd.Document.TransitionRules[5].RuleKey))
	assert.Len(t, cd.Document.TransitionRules, 5)

	// delete non existent rule
	assert.Nil(t, cd.deleteRule(utils.RandomSlice(32)))
	assert.Len(t, cd.Document.TransitionRules, 5)
}

func TestCoreDocument_DeleteTransitionRule(t *testing.T) {
	cd, rule1, role := setupRules(t)

	// add new rule with same role
	assert.False(t, roleNotExists(cd, role))
	assert.Len(t, cd.Document.TransitionRules, 8)
	key, err := AttrKeyFromLabel("test2")
	assert.NoError(t, err)
	rule2, err := cd.AddTransitionRuleForAttribute(role, key)
	assert.NoError(t, err)
	assert.False(t, roleNotExists(cd, role))
	assert.Len(t, cd.Document.TransitionRules, 9)

	// delete rule1
	assert.NoError(t, cd.DeleteTransitionRule(rule1.RuleKey))
	assert.False(t, roleNotExists(cd, role))
	assert.Len(t, cd.Document.TransitionRules, 8)

	// delete rule2
	assert.NoError(t, cd.DeleteTransitionRule(rule2.RuleKey))
	assert.True(t, roleNotExists(cd, role))
	assert.Len(t, cd.Document.TransitionRules, 7)

	// no rule exists
	err = cd.DeleteTransitionRule(rule2.RuleKey)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrTransitionRuleMissing, err))
	assert.True(t, roleNotExists(cd, role))
}

func TestCoreDocument_AddComputeFieldsRule(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// invalid wasm
	wasm := wasmLoader(t, "../testingutils/compute_fields/without_allocate.wasm")
	rules, err := cd.AddComputeFieldsRule(wasm, nil, "")
	assert.Error(t, err)
	assert.Nil(t, rules)
	assert.Len(t, cd.GetComputeFieldsRules(), 0)

	// invalid attribute labels
	wasm = wasmLoader(t, "../testingutils/compute_fields/simple_average.wasm")
	rules, err = cd.AddComputeFieldsRule(wasm, nil, "result")
	assert.Error(t, err)
	assert.Nil(t, rules)
	assert.Len(t, cd.GetComputeFieldsRules(), 0)

	rules, err = cd.AddComputeFieldsRule(wasm, []string{""}, "result")
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEmptyAttrLabel, err))
	assert.Nil(t, rules)
	assert.Len(t, cd.GetComputeFieldsRules(), 0)

	rules, err = cd.AddComputeFieldsRule(wasm, []string{"test"}, "")
	assert.Error(t, err)
	assert.Nil(t, rules)
	assert.Len(t, cd.GetComputeFieldsRules(), 0)

	// add a compute fields rule
	rules, err = cd.AddComputeFieldsRule(wasm, []string{"test"}, "result")
	assert.NoError(t, err)
	assert.Len(t, cd.GetComputeFieldsRules(), 1)
	assert.Equal(t, rules, cd.GetComputeFieldsRules()[0])
}

func TestCoreDocument_CollaboratorCanUpdate(t *testing.T) {
	doc, _, id2, docType := prepareDocument(t)
	wasm := wasmLoader(t, "../testingutils/compute_fields/simple_average.wasm")
	_, err := doc.AddComputeFieldsRule(wasm, []string{"test", "test2", "test3"}, "result")
	assert.NoError(t, err)

	// id2 has write access to only `test`, `test1`, `test2` attributes but not to `result` that is generated
	role, err := doc.AddRole("underwriter", []identity.DID{id2})
	assert.NoError(t, err)
	attrs := getValidComputeFieldAttrs(t)
	for _, attr := range attrs {
		_, err = doc.AddTransitionRuleForAttribute(role.RoleKey, attr.Key)
		assert.NoError(t, err)
	}

	// this should add new
	err = doc.ExecuteComputeFields(computeFieldsTimeout)
	assert.NoError(t, err)
	assert.Len(t, doc.Attributes, 1)
	var key AttrKey
	var result []byte
	for k, v := range doc.Attributes {
		key = k
		result = v.Value.Bytes
	}

	// id2 updates the document
	ndoc, err := doc.PrepareNewVersion([]byte(docType), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	// check should go through
	err = doc.CollaboratorCanUpdate(ndoc, id2, docType)
	assert.NoError(t, err)

	// id2 adds new all the three attributes, which will result in change in the target field
	// even though id2 has no access to that attribute, and change to that doc will not be an issue
	attrDoc, err := ndoc.PrepareNewVersion([]byte(docType), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	attrDoc, err = attrDoc.AddAttributes(CollaboratorsAccess{}, false, nil, getValidComputeFieldAttrs(t)...)
	assert.NoError(t, err)

	// simulate compute fields run
	err = attrDoc.ExecuteComputeFields(computeFieldsTimeout)
	assert.NoError(t, err)
	assert.Len(t, attrDoc.Attributes, 4)
	attr, err := attrDoc.GetAttribute(key)
	assert.NoError(t, err)
	assert.False(t, bytes.Equal(result, attr.Value.Bytes))
	assert.True(t, bytes.Equal(attr.Value.Bytes, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x7, 0xd0}))

	// check should go through
	err = doc.CollaboratorCanUpdate(ndoc, id2, docType)
	assert.NoError(t, err)
}
