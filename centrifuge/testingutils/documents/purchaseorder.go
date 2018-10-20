// +build integration unit

package testingdocuments

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func CreatePOData() purchaseorderpb.PurchaseOrderData {
	return purchaseorderpb.PurchaseOrderData{
		Recipient:   utils.RandomSlice(identity.CentIDLength),
		OrderAmount: 42,
	}
}

func CreateCDWithEmbeddedPO(t *testing.T, poData purchaseorderpb.PurchaseOrderData) *coredocumentpb.CoreDocument {
	identifier := []byte("1")
	poSalt := purchaseorderpb.PurchaseOrderDataSalts{}

	serializedPO, err := proto.Marshal(&poData)
	assert.Nil(t, err, "Could not serialize PurchaseOrderData")

	serializedSalts, err := proto.Marshal(&poSalt)
	assert.Nil(t, err, "Could not serialize PurchaseOrderSalt")

	poAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   serializedPO,
	}
	poAnySalt := any.Any{
		TypeUrl: documenttypes.PurchaseOrderSaltsTypeUrl,
		Value:   serializedSalts,
	}
	coreDocument := &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		EmbeddedData:       &poAny,
		EmbeddedDataSalts:  &poAnySalt,
	}
	return coreDocument
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
