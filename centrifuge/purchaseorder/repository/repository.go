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
	Store(doc *purchaseorderpb.PurchaseOrderDocument) (err error)
	StoreOnce(doc *purchaseorderpb.PurchaseOrderDocument) (err error)
}