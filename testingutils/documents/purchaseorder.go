// +build integration unit

package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/utils"
)

func CreatePOData() purchaseorderpb.PurchaseOrderData {
	return purchaseorderpb.PurchaseOrderData{
		Recipient:   utils.RandomSlice(identity.CentIDLength),
		OrderAmount: 42,
	}
}

func CreatePOPayload() *clientpurchaseorderpb.PurchaseOrderCreatePayload {
	return &clientpurchaseorderpb.PurchaseOrderCreatePayload{
		Data: &clientpurchaseorderpb.PurchaseOrderData{
			Recipient:   "0x010203040506",
			OrderAmount: 42,
			ExtraData:   "0x01020302010203",
			Currency:    "EUR",
		},
		Collaborators: []string{"0x010101010101"},
	}
}
