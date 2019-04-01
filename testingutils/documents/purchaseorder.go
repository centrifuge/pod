// +build integration unit testworld

package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
)

func CreatePOData() purchaseorderpb.PurchaseOrderData {
	recipient := testingidentity.GenerateRandomDID()
	return purchaseorderpb.PurchaseOrderData{
		Recipient:   recipient[:],
		TotalAmount: []byte{0, 42, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func CreatePOPayload() *clientpurchaseorderpb.PurchaseOrderCreatePayload {
	return &clientpurchaseorderpb.PurchaseOrderCreatePayload{
		Data: &clientpurchaseorderpb.PurchaseOrderData{
			Recipient:   "0xEA939D5C0494b072c51565b191eE59B5D34fbf79",
			TotalAmount: "42",
			Currency:    "EUR",
		},
		WriteAccess: &documentpb.WriteAccess{Collaborators: []string{testingidentity.GenerateRandomDID().String()}},
	}
}
