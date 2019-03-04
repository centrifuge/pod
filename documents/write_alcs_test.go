// +build unit

package documents

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestWriteACLs_getChangedFields_different_types(t *testing.T) {
	ocd := newCoreDocument().Document
	ncd := invoicepb.InvoiceData{
		Currency: "EUR",
	}

	oldTree := getTree(t, &ocd)
	newTree := getTree(t, &ncd)

	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	// cf length should be len(ocd) and len(ncd) = 28 changed field
	assert.Len(t, cf, 28)

}

func TestWriteACLs_getChangedFields_same_document(t *testing.T) {
	ocd := newCoreDocument().Document
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

func TestWriteACLs_getChangedFields_with_core_document(t *testing.T) {
	doc := newCoreDocument()
	doc.Document.DocumentRoot = utils.RandomSlice(32)
	ndoc, err := doc.PrepareNewVersion([]string{testingidentity.GenerateRandomDID().String()}, true)
	assert.NoError(t, err)

	// preparing new version would have changed the following properties
	// current_version
	// previous_version
	// next_version
	// previous_root
	// roles
	// read_rules.roles
	// read_rules.action
	oldTree := getTree(t, &doc.Document)
	newTree := getTree(t, &ndoc.Document)
	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 7)

	// prepare new version with out new collaborators
	// this should only change
	// current_version
	// previous_version
	// next_version
	// previous_root
	doc = ndoc
	doc.Document.DocumentRoot = utils.RandomSlice(32)
	ndoc, err = doc.PrepareNewVersion(nil, true)
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document)
	newTree = getTree(t, &ndoc.Document)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 4)

	// test with different document
	// this will change
	// document identifier
	// current version
	// previous version
	// next version
	// previous_root
	// roles (new doc will have empty role while old one has one role)
	// read_rules (new doc will have empty read_rules while old one has read_rules)
	doc = ndoc
	ndoc = newCoreDocument()
	oldTree = getTree(t, &doc.Document)
	newTree = getTree(t, &ndoc.Document)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 8)

	// add different roles and read rules and check
	ndoc.Document.DocumentRoot = utils.RandomSlice(32)
	ndoc, err = ndoc.PrepareNewVersion([]string{testingidentity.GenerateRandomDID().String()}, true)
	assert.NoError(t, err)
	oldTree = getTree(t, &doc.Document)
	newTree = getTree(t, &ndoc.Document)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 8)
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
	dueDate, err = ptypes.TimestampProto(time.Now().Add(30 * time.Minute))
	assert.NoError(t, err)
	ndoc := &invoicepb.InvoiceData{
		InvoiceNumber: doc.InvoiceNumber,
		SenderName:    doc.SenderName,
		RecipientName: doc.RecipientName,
		DateCreated:   doc.DateCreated,
		DueDate:       dueDate, // updated
		Currency:      "EUR",   // new field
	}

	oldTree = getTree(t, doc)
	newTree = getTree(t, ndoc)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 2)

	// completely new doc
	// this should give 5 property changes
	ndoc = new(invoicepb.InvoiceData)
	oldTree = getTree(t, doc)
	newTree = getTree(t, ndoc)
	cf = getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	assert.Len(t, cf, 5)
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
