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
	FindById(id []byte) (inv *purchaseorderpb.PurchaseOrderDocument, err error)
	Store(inv *purchaseorderpb.PurchaseOrderDocument) (err error)
}
