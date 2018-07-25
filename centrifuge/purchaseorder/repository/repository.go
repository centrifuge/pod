package purchaseorderrepository

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
)

var purchaseorderRepository PurchaseOrderRepository

func GetPurchaseOrderRepository() PurchaseOrderRepository {
	return purchaseorderRepository
}

type PurchaseOrderRepository interface {
	GetKey(id []byte) ([]byte)
	FindById(id []byte) (doc *purchaseorderpb.PurchaseOrderDocument, err error)

	// CreateOrUpdate functions similar to a REST HTTP PUT where the document is either created or updated regardless if it existed before
	CreateOrUpdate(doc *purchaseorderpb.PurchaseOrderDocument) (err error)

	// Create will only create a document initially. If the same document (as identified by its DocumentIdentifier) exists
	// the Create method will error out
	Create(doc *purchaseorderpb.PurchaseOrderDocument) (err error)
}