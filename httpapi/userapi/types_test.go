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
		Data: &transferdetails.TransferDetailData{
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
		Data: &transferdetails.TransferDetailData{
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
