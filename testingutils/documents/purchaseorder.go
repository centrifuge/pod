// +build integration unit testworld

package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
)

func CreatePOData() purchaseorderpb.PurchaseOrderData {
	recipient := testingidentity.GenerateRandomDID()
	return purchaseorderpb.PurchaseOrderData{
		Recipient:   recipient[:],
		OrderAmount: []byte{0, 42, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func CreatePOPayload() *clientpurchaseorderpb.PurchaseOrderCreatePayload {
	return &clientpurchaseorderpb.PurchaseOrderCreatePayload{
		Data: &clientpurchaseorderpb.PurchaseOrderData{
			Recipient:   "0xea939d5c0494b072c51565b191ee59b5d34fbf79",
			OrderAmount: "42",
			ExtraData:   "0x01020302010203",
			Currency:    "EUR",
		},
		Collaborators: []string{testingidentity.GenerateRandomDID().String()},
	}
}
