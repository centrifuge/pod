package purchaseorder

import (
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("purchaseorder")

type PurchaseOrder struct {
	Document *purchaseorderpb.PurchaseOrderDocument
}

func NewPurchaseOrder(poDoc *purchaseorderpb.PurchaseOrderDocument) *PurchaseOrder {
	order := &PurchaseOrder{poDoc}
	// IF salts have not been provided, let's generate them
	if poDoc.Salts == nil {
		purchaseorderSalts := purchaseorderpb.PurchaseOrderDataSalts{}
		proofs.FillSalts(&purchaseorderSalts)
		order.Document.Salts = &purchaseorderSalts
	}
	return order
}

func NewEmptyPurchaseOrder() *PurchaseOrder {
	purchaseorderSalts := purchaseorderpb.PurchaseOrderDataSalts{}
	proofs.FillSalts(&purchaseorderSalts)
	doc := purchaseorderpb.PurchaseOrderDocument{
		CoreDocument: &coredocumentpb.CoreDocument{},
		Data:         &purchaseorderpb.PurchaseOrderData{},
		Salts:        &purchaseorderSalts,
	}
	return &PurchaseOrder{&doc}
}

func NewPurchaseOrderFromCoreDocument(coredocument *coredocumentpb.CoreDocument) (order *PurchaseOrder) {
	if coredocument.EmbeddedData.TypeUrl != documenttypes.PurchaseOrderDataTypeUrl ||
		coredocument.EmbeddedDataSalts.TypeUrl != documenttypes.PurchaseOrderSaltsTypeUrl {
		log.Fatal("Trying to convert document with incorrect schema")
	}

	purchaseorderData := &purchaseorderpb.PurchaseOrderData{}
	proto.Unmarshal(coredocument.EmbeddedData.Value, purchaseorderData)

	purchaseorderSalts := &purchaseorderpb.PurchaseOrderDataSalts{}
	proto.Unmarshal(coredocument.EmbeddedDataSalts.Value, purchaseorderSalts)

	emptiedCoreDoc := coredocumentpb.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredocument)
	emptiedCoreDoc.EmbeddedData = nil
	emptiedCoreDoc.EmbeddedDataSalts = nil
	order = NewEmptyPurchaseOrder()
	order.Document.Data = purchaseorderData
	order.Document.Salts = purchaseorderSalts
	order.Document.CoreDocument = &emptiedCoreDoc
	return
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

func (order *PurchaseOrder) CalculateMerkleRoot() error {
	tree, err := order.getDocumentTree()
	if err != nil {
		return err
	}
	// TODO: below should actually be stored as CoreDocument.DataRoot
	order.Document.CoreDocument.DocumentRoot = tree.RootHash()
	return nil
}

func (order *PurchaseOrder) CreateProofs(fields []string) (proofs []*proofs.Proof, err error) {
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
	return
}

func (order *PurchaseOrder) ConvertToCoreDocument() (coredocpb *coredocumentpb.CoreDocument) {
	coredocpb = &coredocumentpb.CoreDocument{}
	proto.Merge(coredocpb, order.Document.CoreDocument)
	serializedPurchaseOrder, err := proto.Marshal(order.Document.Data)
	if err != nil {
		log.Fatalf("Could not serialize PurchaseOrderData: %s", err)
	}

	purchaseorderAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   serializedPurchaseOrder,
	}

	serializedSalts, err := proto.Marshal(order.Document.Salts)
	if err != nil {
		log.Fatalf("Could not serialize PurchaseOrderSalts: %s", err)
	}

	purchaseorderSaltsAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coredocpb.EmbeddedData = &purchaseorderAny
	coredocpb.EmbeddedDataSalts = &purchaseorderSaltsAny
	return coredocpb
}
