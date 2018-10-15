// Important
// Note: After the migration to the new invoice model this file will not exist anymore

package invoice

import (
	"crypto/sha256"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("invoice")

// Invoice is a wrapper for invoice protobuf
// Deprecated: in favour of implementation of documents.Model interface (InvoiceModel).
// TODO remove
type Invoice struct {
	Document *invoicepb.InvoiceDocument
}

// Wrap wraps the protobuf invoice within Invoice
func Wrap(invDoc *invoicepb.InvoiceDocument) (*Invoice, error) {
	if invDoc == nil {
		return nil, centerrors.NilError(invDoc)
	}
	return &Invoice{invDoc}, nil
}

// New returns a new Invoice with salts, merkle root, and coredocument generated
// Deprecated
func New(invDoc *invoicepb.InvoiceDocument, collaborators [][]byte) (*Invoice, error) {
	inv, err := Wrap(invDoc)
	if err != nil {
		return nil, err
	}
	// IF salts have not been provided, let's generate them
	if invDoc.Salts == nil {
		invoiceSalts := invoicepb.InvoiceDataSalts{}
		proofs.FillSalts(invDoc.Data, &invoiceSalts)
		invDoc.Salts = &invoiceSalts
	}

	if inv.Document.CoreDocument == nil {
		inv.Document.CoreDocument = coredocument.New()
		inv.Document.CoreDocument.EmbeddedData = &any.Any{TypeUrl: documenttypes.InvoiceDataTypeUrl, Value: []byte{}}
	}
	inv.Document.CoreDocument.Collaborators = collaborators
	coredocument.FillSalts(inv.Document.CoreDocument)

	err = inv.CalculateMerkleRoot()
	if err != nil {
		return nil, err
	}

	return inv, nil
}

// Empty returns an empty invoice
func Empty() *Invoice {
	invoiceSalts := invoicepb.InvoiceDataSalts{}
	proofs.FillSalts(&invoicepb.InvoiceData{}, &invoiceSalts)
	doc := invoicepb.InvoiceDocument{
		CoreDocument: &coredocumentpb.CoreDocument{},
		Data:         &invoicepb.InvoiceData{},
		Salts:        &invoiceSalts,
	}
	return &Invoice{&doc}
}

// NewFromCoreDocument returns an Invoice from Core Document
// Will Empty embedded fields as they are represented as data in the invoice header
func NewFromCoreDocument(coreDocument *coredocumentpb.CoreDocument) (*Invoice, error) {
	if coreDocument == nil {
		return nil, centerrors.NilError(coreDocument)
	}
	if coreDocument.EmbeddedData.TypeUrl != documenttypes.InvoiceDataTypeUrl ||
		coreDocument.EmbeddedDataSalts.TypeUrl != documenttypes.InvoiceSaltsTypeUrl {
		return nil, fmt.Errorf("trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	proto.Unmarshal(coreDocument.EmbeddedData.Value, invoiceData)

	invoiceSalts := &invoicepb.InvoiceDataSalts{}
	proto.Unmarshal(coreDocument.EmbeddedDataSalts.Value, invoiceSalts)

	emptiedCoreDoc := coredocumentpb.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coreDocument)
	emptiedCoreDoc.EmbeddedData = nil
	emptiedCoreDoc.EmbeddedDataSalts = nil
	inv := Empty()
	inv.Document.Data = invoiceData
	inv.Document.Salts = invoiceSalts
	inv.Document.CoreDocument = &emptiedCoreDoc
	return inv, nil
}

func (inv *Invoice) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: sha256.New()})
	err = t.AddLeavesFromDocument(inv.Document.Data, inv.Document.Salts)
	if err != nil {
		log.Error("getDocumentDataTree:", err)
		return nil, err
	}
	err = t.Generate()
	if err != nil {
		log.Error("getDocumentDataTree:", err)
		return nil, err
	}
	return &t, nil
}

// CalculateMerkleRoot calculates the invoice merkle root
// TODO: this method is a dangerous one. Generating the different roots shouldn't be done in one step (lucas)
func (inv *Invoice) CalculateMerkleRoot() error {
	tree, err := inv.getDocumentDataTree()
	if err != nil {
		return err
	}
	inv.Document.CoreDocument.DataRoot = tree.RootHash()
	err = coredocument.CalculateSigningRoot(inv.Document.CoreDocument)
	return err
}

// CreateProofs generates proofs for given fields
func (inv *Invoice) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := inv.getDocumentDataTree()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return coredocument.CreateProofs(tree, inv.Document.CoreDocument, fields)
}

// ConvertToCoreDocument converts invoice document to coredocument
func (inv *Invoice) ConvertToCoreDocument() (coredocpb *coredocumentpb.CoreDocument, err error) {
	coredocpb = new(coredocumentpb.CoreDocument)
	proto.Merge(coredocpb, inv.Document.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Document.Data)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise InvoiceData")
	}

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}

	serializedSalts, err := proto.Marshal(inv.Document.Salts)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise InvoiceSalts")
	}

	invoiceSaltsAny := any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coredocpb.EmbeddedData = &invoiceAny
	coredocpb.EmbeddedDataSalts = &invoiceSaltsAny
	return
}

// Validate validates the invoice document
func Validate(doc *invoicepb.InvoiceDocument) error {
	if doc == nil {
		return fmt.Errorf("nil document")
	}

	if err := coredocument.Validate(doc.CoreDocument); err != nil {
		return err
	}

	if doc.Data == nil {
		return fmt.Errorf("missing invoice data")
	}

	// checking for nil salts should be okay for now
	// once the converters are in, salts will be filled during conversion
	if doc.Salts == nil {
		return fmt.Errorf("missing invoice salts")
	}

	return nil
}
