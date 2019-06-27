// +build unit

package userapi

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/stretchr/testify/assert"
)

func TestTypes_toTransferDetailCreatePayload(t *testing.T) {
	req := CreateTransferDetailRequest{
		DocumentID: "test_id",
		Data: transferdetails.Data{
			TransferID: "test",
		},
	}
	// success
	payload, err := toTransferDetailCreatePayload(req)
	assert.NoError(t, err)
	assert.Equal(t, payload.DocumentID, "test_id")
	assert.NotNil(t, payload.Data)
	assert.Equal(t, payload.Data.TransferID, req.Data.TransferID)
}

func TestTypes_toTransferDetailUpdatePayload(t *testing.T) {
	req := UpdateTransferDetailRequest{
		DocumentID: "test_id",
		TransferID: "transfer_id",
		Data: transferdetails.Data{
			TransferID: "test",
		},
	}
	// success
	payload, err := toTransferDetailUpdatePayload(req)
	assert.NoError(t, err)
	assert.Equal(t, "test_id", payload.DocumentID)
	assert.Equal(t, "transfer_id", payload.TransferID)
	assert.NotNil(t, payload.Data)
	assert.Equal(t, payload.Data.TransferID, req.Data.TransferID)
}

func transferData() map[string]interface{} {
	return map[string]interface{}{
		"status":       "unpaid",
		"amount":       "300",
		"sender_id":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"recipient_id": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"currency":     "EUR",
	}
}
