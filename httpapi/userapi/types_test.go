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
		Data: transferdetails.TransferDetailData{
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
		Data: transferdetails.TransferDetailData{
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

func invoiceData() map[string]interface{} {
	return map[string]interface{}{
		"number":       "12345",
		"status":       "unpaid",
		"gross_amount": "12.345",
		"recipient":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_paid":    "2019-05-24T14:48:44Z",        // rfc3339
		"currency":     "EUR",
		"attachments": []map[string]interface{}{
			{
				"name":      "test",
				"file_type": "pdf",
				"size":      1000202,
				"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
			},
		},
	}
}
