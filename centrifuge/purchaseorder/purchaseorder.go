package purchaseorder

import (
	"crypto/sha256"
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("purchaseorder")

type PurchaseOrder struct {
	Document *purchaseorderpb.PurchaseOrderDocument
}

// Wrap wraps the purchase order protobuf inside PurchaseOrder
func Wrap(poDoc *purchaseorderpb.PurchaseOrderDocument) (*PurchaseOrder, error) {
	if poDoc == nil {
		return nil, errors.NilError(poDoc)
	}
	return &PurchaseOrder{poDoc}, nil
}

// New generates a new purchase order and generate salts, merkle root and coredocument
func New(poDoc *purchaseorderpb.PurchaseOrderDocument) (*PurchaseOrder, error) {
	po, err := Wrap(poDoc)
	if err != nil {
		return nil, err
	}

	// IF salts have not been provided, let's generate them
	if poDoc.Salts == nil {
		salts := purchaseorderpb.PurchaseOrderDataSalts{}
		proofs.FillSalts(&salts)
		po.Document.Salts = &salts
	}

	if po.Document.CoreDocument == nil {
		po.Document.CoreDocument = coredocument.New()
	}

	err = po.CalculateMerkleRoot()
	if err != nil {
		return nil, err
	}

	return po, nil
}

// Empty returns an empty purchase order with salts filled
func Empty() *PurchaseOrder {
	salts := purchaseorderpb.PurchaseOrderDataSalts{}
	proofs.FillSalts(&salts)
	doc := purchaseorderpb.PurchaseOrderDocument{
		CoreDocument: &coredocumentpb.CoreDocument{},
		Data:         &purchaseorderpb.PurchaseOrderData{},
		Salts:        &salts,
	}
	return &PurchaseOrder{&doc}
}

// NewFromCoreDocument unmarshalls invoice from coredocument
func NewFromCoreDocument(coredocument *coredocumentpb.CoreDocument) (*PurchaseOrder, error) {
	if coredocument == nil {
		return nil, errors.NilError(coredocument)
	}

	if coredocument.EmbeddedData.TypeUrl != documenttypes.PurchaseOrderDataTypeUrl ||
		coredocument.EmbeddedDataSalts.TypeUrl != documenttypes.PurchaseOrderSaltsTypeUrl {
		return nil, fmt.Errorf("trying to convert document with incorrect schema")
	}

	poData := &purchaseorderpb.PurchaseOrderData{}
	proto.Unmarshal(coredocument.EmbeddedData.Value, poData)

	poSalts := &purchaseorderpb.PurchaseOrderDataSalts{}
	proto.Unmarshal(coredocument.EmbeddedDataSalts.Value, poSalts)

	emptiedCoreDoc := coredocumentpb.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredocument)
	emptiedCoreDoc.EmbeddedData = nil
	emptiedCoreDoc.EmbeddedDataSalts = nil
	order := Empty()
	order.Document.Data = poData
	order.Document.Salts = poSalts
	order.Document.CoreDocument = &emptiedCoreDoc
	return order, nil
}

func (order *PurchaseOrder) getDocumentTree() (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree()
	sha256Hash := sha256.New()
	t.SetHashFunc(sha256Hash)
	err = t.FillTree(order.Document.Data, order.Document.Salts)
	if err != nil {
		log.Error("getDocumentTree:", err)
		return nil, err
	}
	return &t, nil
}

// CalculateMerkleRoot calculates MerkleRoot of the document
func (order *PurchaseOrder) CalculateMerkleRoot() error {
	tree, err := order.getDocumentTree()
	if err != nil {
		return err
	}
	order.Document.CoreDocument.DataRoot = tree.RootHash()
	return nil
}

// CreateProofs generates proofs for given fields
func (order *PurchaseOrder) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := order.getDocumentTree()
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

	return proofs, nil
}

// ConvertToCoreDocument converts purchaseOrder to a core document
func (order *PurchaseOrder) ConvertToCoreDocument() (coredocpb *coredocumentpb.CoreDocument, err error) {
	coredocpb = &coredocumentpb.CoreDocument{}
	proto.Merge(coredocpb, order.Document.CoreDocument)
	serializedPurchaseOrder, err := proto.Marshal(order.Document.Data)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't serialize PurchaseOrderData")
	}

	po := any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   serializedPurchaseOrder,
	}

	serializedSalts, err := proto.Marshal(order.Document.Salts)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't serialize PurchaseOrderSalts")
	}

	poSalts := any.Any{
		TypeUrl: documenttypes.PurchaseOrderSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coredocpb.EmbeddedData = &po
	coredocpb.EmbeddedDataSalts = &poSalts
	return coredocpb, nil
}

// Validate validates the purchase order document
func Validate(doc *purchaseorderpb.PurchaseOrderDocument) (valid bool, errMsg string, errs map[string]string) {
	if doc == nil {
		return false, errors.NilDocument, nil
	}

	if valid, errMsg, errs = coredocument.Validate(doc.CoreDocument); !valid {
		return valid, errMsg, errs
	}

	if doc.Data == nil {
		return false, errors.NilDocumentData, nil
	}

	data := doc.Data
	errs = make(map[string]string)

	// ideally these check should be done in the client purchase order
	// once the converters are done, we can move the following checks there
	if data.PoNumber == "" {
		errs["po_number"] = errors.RequiredField
	}

	if data.OrderName == "" {
		errs["po_order_name"] = errors.RequiredField
	}

	if data.OrderZipcode == "" {
		errs["po_order_zip_code"] = errors.RequiredField
	}

	// for now, mandating at least one character
	if data.OrderCountry == "" {
		errs["po_order_country"] = errors.RequiredField
	}

	if data.RecipientName == "" {
		errs["po_recipient_name"] = errors.RequiredField
	}

	if data.RecipientZipcode == "" {
		errs["po_recipient_zip_code"] = errors.RequiredField
	}

	if data.RecipientCountry == "" {
		errs["po_recipient_country"] = errors.RequiredField
	}

	if data.Currency == "" {
		errs["po_currency"] = errors.RequiredField
	}

	if data.OrderAmount <= 0 {
		errs["po_order_amount"] = errors.RequirePositiveNumber
	}

	// checking for nil salts should be okay for now
	// once the converters are in, salts will be filled during conversion
	if doc.Salts == nil {
		errs["po_salts"] = errors.RequiredField
	}

	if len(errs) < 1 {
		return true, "", nil
	}

	return false, "Invalid Purchase Order", errs
}
