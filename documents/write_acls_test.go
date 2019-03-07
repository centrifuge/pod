// +build unit

package documents

import (
	"crypto/sha256"
	"reflect"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestWriteACLs_getChangedFields_different_types(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	ocd := cd.Document
	ncd := invoicepb.InvoiceData{
		Currency: "EUR",
	}

	oldTree := getTree(t, &ocd)
	newTree := getTree(t, &ncd)

	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	// cf length should be len(ocd) and len(ncd) = 30 changed field
	assert.Len(t, cf, 30)

}

func TestWriteACLs_getChangedFields_same_document(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	ocd := cd.Document
	oldTree := getTree(t, &ocd)
	newTree := getTree(t, &ocd)
	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 0)

	// check hashed field
	ocd.PreviousRoot = utils.RandomSlice(32)
	oldTree = getTree(t, &ocd)
	newTree = getTree(t, &ocd)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 0)
}

func testExpectedProps(t *testing.T, cf []changedField, eprops map[string]struct{}) {
	for _, f := range cf {
		_, ok := eprops[hexutil.Encode(f.property)]
		if !ok {
			assert.Failf(t, "", "expected %x property to be present", f.property)
		}
	}
}

func TestWriteACLs_getChangedFields_with_core_document(t *testing.T) {
	doc, err := newCoreDocument()
	assert.NoError(t, err)
	doc.Document.DocumentRoot = utils.RandomSlice(32)
	ndoc, err := doc.PrepareNewVersion([]string{testingidentity.GenerateRandomDID().String()}, true)
	assert.NoError(t, err)

	// preparing new version would have changed the following properties
	// current_version
	// previous_version
	// next_version
	// previous_root
	// roles (x2)
	// current pre image
	// next pre image
	// read_rules.roles
	// read_rules.action
	// transition_rules.RuleKey
	// transition_rules.Roles
	// transition_rules.MatchType
	// transition_rules.Action
	// transition_rules.Field
	oldTree := getTree(t, &doc.Document)
	newTree := getTree(t, &ndoc.Document)
	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 15)
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
	doc.Document.DocumentRoot = utils.RandomSlice(32)
	ndoc, err = doc.PrepareNewVersion(nil, true)
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document)
	newTree = getTree(t, &ndoc.Document)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 6)
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
	// transition_rules (new doc will have empty transition_rules while old one has transition_rules)
	doc = ndoc
	ndoc, err = newCoreDocument()
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document)
	newTree = getTree(t, &ndoc.Document)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 16)
	rprop = append(doc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop2 = append(doc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	eprops = map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}):                         {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5}):                         {},
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
	ndoc.Document.DocumentRoot = utils.RandomSlice(32)
	ndoc, err = ndoc.PrepareNewVersion([]string{testingidentity.GenerateRandomDID().String()}, true)
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document)
	newTree = getTree(t, &ndoc.Document)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 14)
	rprop = append(doc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop2 = append(doc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop3 := append(ndoc.Document.Roles[0].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	rprop4 := append(ndoc.Document.Roles[1].RoleKey, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0)
	eprops = map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 9}):                                      {},
		hexutil.Encode([]byte{0, 0, 0, 4}):                                      {},
		hexutil.Encode([]byte{0, 0, 0, 3}):                                      {},
		hexutil.Encode([]byte{0, 0, 0, 16}):                                     {},
		hexutil.Encode([]byte{0, 0, 0, 2}):                                      {},
		hexutil.Encode([]byte{0, 0, 0, 22}):                                     {},
		hexutil.Encode([]byte{0, 0, 0, 23}):                                     {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}): {},
		hexutil.Encode([]byte{0, 0, 0, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode([]byte{0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}): {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop...)):                                            {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop2...)):                                           {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop3...)):                                           {},
		hexutil.Encode(append([]byte{0, 0, 0, 1}, rprop4...)):                                           {},
	}
	testExpectedProps(t, cf, eprops)
}

func TestWriteACLs_getChangedFields_invoice_document(t *testing.T) {
	dueDate, err := ptypes.TimestampProto(time.Now().Add(10 * time.Minute))
	assert.NoError(t, err)

	// no change
	doc := &invoicepb.InvoiceData{
		InvoiceNumber: "12345",
		SenderName:    "Alice",
		RecipientName: "Bob",
		DateCreated:   ptypes.TimestampNow(),
		DueDate:       dueDate,
	}

	oldTree := getTree(t, doc)
	newTree := getTree(t, doc)
	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 0)

	// updated doc
	ndoc := &invoicepb.InvoiceData{
		InvoiceNumber: "123456", // updated
		SenderName:    doc.SenderName,
		RecipientName: doc.RecipientName,
		DateCreated:   doc.DateCreated,
		DueDate:       doc.DueDate,
		Currency:      "EUR", // new field
	}

	oldTree = getTree(t, doc)
	newTree = getTree(t, ndoc)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 2)
	eprops := map[string]changedField{
		hexutil.Encode([]byte{0, 0, 0, 1}): {
			property: []byte{0, 0, 0, 1},
			old:      []byte{49, 50, 51, 52, 53},
			new:      []byte{49, 50, 51, 52, 53, 54},
		},

		hexutil.Encode([]byte{0, 0, 0, 13}): {
			property: []byte{0, 0, 0, 13},
			old:      []byte{},
			new:      []byte{69, 85, 82},
		},
	}

	for _, f := range cf {
		ef, ok := eprops[hexutil.Encode(f.property)]
		if !ok {
			t.Fatalf("expected %x property change", f.property)
		}

		assert.True(t, reflect.DeepEqual(f, ef))
	}

	// completely new doc
	// this should give 5 property changes
	ndoc = new(invoicepb.InvoiceData)
	oldTree = getTree(t, doc)
	newTree = getTree(t, ndoc)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 5)
	eprps := map[string]struct{}{
		hexutil.Encode([]byte{0, 0, 0, 1}):  {},
		hexutil.Encode([]byte{0, 0, 0, 3}):  {},
		hexutil.Encode([]byte{0, 0, 0, 8}):  {},
		hexutil.Encode([]byte{0, 0, 0, 23}): {},
		hexutil.Encode([]byte{0, 0, 0, 22}): {},
	}
	testExpectedProps(t, cf, eprps)
}

func getTree(t *testing.T, doc proto.Message) *proofs.DocumentTree {
	tr := proofs.NewDocumentTree(proofs.TreeOptions{
		CompactProperties: true,
		EnableHashSorting: true,
		SaltsLengthSuffix: proofs.DefaultSaltsLengthSuffix,
		Hash:              sha256.New(),
	})

	tree := &tr
	assert.NoError(t, tree.AddLeavesFromDocument(doc))
	assert.NoError(t, tree.Generate())
	return tree
}

func TestReadACLs_initTransitionRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.initTransitionRules(nil)
	assert.Nil(t, cd.Document.Roles)
	assert.Nil(t, cd.Document.TransitionRules)

	collab := []identity.DID{testingidentity.GenerateRandomDID()}
	cd.initTransitionRules(collab)
	assert.Len(t, cd.Document.TransitionRules, 1)
	assert.Len(t, cd.Document.Roles, 1)

	cd.initTransitionRules(collab)
	assert.Len(t, cd.Document.TransitionRules, 1)
	assert.Len(t, cd.Document.Roles, 1)
}
