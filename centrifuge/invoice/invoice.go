package invoice

import (
	"crypto/sha256"
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("invoice")

type Invoice struct {
	Document *invoicepb.InvoiceDocument
}

func NewInvoice(invDoc *invoicepb.InvoiceDocument) (*Invoice, error) {
	if invDoc == nil {
		return nil, errors.NilError(invDoc)
	}
	inv := &Invoice{invDoc}
	// IF salts have not been provided, let's generate them
	if invDoc.Salts == nil {
		invoiceSalts := invoicepb.InvoiceDataSalts{}
		proofs.FillSalts(invDoc.Data, &invoiceSalts)
		inv.Document.Salts = &invoiceSalts
	}
	return inv, nil
}

func NewEmptyInvoice() *Invoice {
	invoiceSalts := invoicepb.InvoiceDataSalts{}
	invoiceData := invoicepb.InvoiceData{}
	proofs.FillSalts(&invoiceData, &invoiceSalts)
	doc := invoicepb.InvoiceDocument{
		CoreDocument: &coredocumentpb.CoreDocument{},
		Data:         &invoiceData,
		Salts:        &invoiceSalts,
	}
	return &Invoice{&doc}
}

func NewInvoiceFromCoreDocument(coreDocument *coredocumentpb.CoreDocument) (*Invoice, error) {
	if coreDocument == nil {
		return nil, errors.NilError(coreDocument)
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
	inv := NewEmptyInvoice()
	inv.Document.Data = invoiceData
	inv.Document.Salts = invoiceSalts
	inv.Document.CoreDocument = &emptiedCoreDoc
	return inv, nil
}

func (inv *Invoice) getDocumentTree() (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree(proofs.TreeOptions{Hash: sha256.New()})
	err = t.AddLeavesFromDocument(inv.Document.Data, inv.Document.Salts)
	if err != nil {
		log.Error("getDocumentTree:", err)
		return nil, err
	}
	err = t.Generate()
	if err != nil {
		log.Error("getDocumentTree:", err)
		return nil, err
	}
	return &t, nil
}

func (inv *Invoice) CalculateMerkleRoot() error {
	tree, err := inv.getDocumentTree()
	if err != nil {
		return err
	}
	inv.Document.CoreDocument.DataRoot = tree.RootHash()
	return nil
}

func (inv *Invoice) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := inv.getDocumentTree()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, field := range fields {
		proof, err := tree.CreateProof(field)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		proofs = append(proofs, &proof)
	}
	return
}

func (inv *Invoice) ConvertToCoreDocument() (coredocpb *coredocumentpb.CoreDocument) {
	coredocpb = &coredocumentpb.CoreDocument{}
	proto.Merge(coredocpb, inv.Document.CoreDocument)
	serializedInvoice, err := proto.Marshal(inv.Document.Data)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceData: %s", err)
	}

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}

	serializedSalts, err := proto.Marshal(inv.Document.Salts)
	if err != nil {
		log.Fatalf("Could not serialize InvoiceSalts: %s", err)
	}

	invoiceSaltsAny := any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coredocpb.EmbeddedData = &invoiceAny
	coredocpb.EmbeddedDataSalts = &invoiceSaltsAny
	return
}
