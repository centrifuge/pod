package purchaseorder

import (
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/documenttypes"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("purchaseorder")

type PurchaseOrder struct {
	Document *purchaseorderpb.PurchaseOrderDocument
}

func NewPurchaseOrder(invDoc *purchaseorderpb.PurchaseOrderDocument) *PurchaseOrder {
	inv := &PurchaseOrder{invDoc}
	// IF salts have not been provided, let's generate them
	if invDoc.Salts == nil {
		purchaseorderSalts := purchaseorderpb.PurchaseOrderDataSalts{}
		proofs.FillSalts(&purchaseorderSalts)
		inv.Document.Salts = &purchaseorderSalts
	}
	return inv
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

func NewPurchaseOrderFromCoreDocument(coredocument *coredocument.CoreDocument) (inv *PurchaseOrder) {
	if coredocument.Document.EmbeddedData.TypeUrl != documenttypes.PurchaseOrderDataTypeUrl ||
		coredocument.Document.EmbeddedDataSalts.TypeUrl != documenttypes.PurchaseOrderSaltsTypeUrl {
		log.Fatal("Trying to convert document with incorrect schema")
	}

	purchaseorderData := &purchaseorderpb.PurchaseOrderData{}
	proto.Unmarshal(coredocument.Document.EmbeddedData.Value, purchaseorderData)

	purchaseorderSalts := &purchaseorderpb.PurchaseOrderDataSalts{}
	proto.Unmarshal(coredocument.Document.EmbeddedDataSalts.Value, purchaseorderSalts)

	emptiedCoreDoc := coredocumentpb.CoreDocument{}
	proto.Merge(&emptiedCoreDoc, coredocument.Document)
	emptiedCoreDoc.EmbeddedData = nil
	emptiedCoreDoc.EmbeddedDataSalts = nil
	inv = NewEmptyPurchaseOrder()
	inv.Document.Data = purchaseorderData
	inv.Document.Salts = purchaseorderSalts
	inv.Document.CoreDocument = &emptiedCoreDoc
	return
}

func (inv *PurchaseOrder) getDocumentTree() (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree()
	sha256Hash := sha256.New()
	t.SetHashFunc(sha256Hash)
	err = t.FillTree(inv.Document.Data, inv.Document.Salts)
	if err != nil {
		log.Error("getDocumentTree:", err)
		return nil, err
	}
	return &t, nil
}

func (inv *PurchaseOrder) CalculateMerkleRoot() error {
	tree, err := inv.getDocumentTree()
	if err != nil {
		return err
	}
	// TODO: below should actually be stored as CoreDocument.DataMerkleRoot
	inv.Document.CoreDocument.DocumentRoot = tree.RootHash()
	return nil
}

func (inv *PurchaseOrder) CreateProofs(fields []string) (proofs []*proofs.Proof, err error) {
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

func (inv *PurchaseOrder) ConvertToCoreDocument() (coredocument coredocument.CoreDocument) {
	coredocpb := &coredocumentpb.CoreDocument{}
	proto.Merge(coredocpb, inv.Document.CoreDocument)
	serializedPurchaseOrder, err := proto.Marshal(inv.Document.Data)
	if err != nil {
		log.Fatalf("Could not serialize PurchaseOrderData: %s", err)
	}

	purchaseorderAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   serializedPurchaseOrder,
	}

	serializedSalts, err := proto.Marshal(inv.Document.Salts)
	if err != nil {
		log.Fatalf("Could not serialize PurchaseOrderSalts: %s", err)
	}

	purchaseorderSaltsAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coredocpb.EmbeddedData = &purchaseorderAny
	coredocpb.EmbeddedDataSalts = &purchaseorderSaltsAny
	coredocument.Document = coredocpb
	return
}
