// nolint
package userapi

import (
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
)

// TODO: think: generic custom attribute set creation?

//CreateTransferDetailRequest is the request body for creating a Transfer Detail
type CreateTransferDetailRequest struct {
	DocumentID string                              `json:"document_id" swaggertype:"primitive,string"`
	Data       *transferdetails.TransferDetailData `json:"data"`
}

// UpdateTransferDetailRequest is the request body for updating a Transfer Detail
type UpdateTransferDetailRequest struct {
	DocumentID string                              `json:"document_id" swaggertype:"primitive,string"`
	TransferID string                              `json:"transfer_id" swaggertype:"primitive,string"`
	Data       *transferdetails.TransferDetailData `json:"data"`
}

// TransferDetailResponse is the response body when fetching a Transfer Detail
type TransferDetailResponse struct {
	Header *coreapi.ResponseHeader             `json:"header"`
	Data   *transferdetails.TransferDetailData `json:"data"`
}

// TransferDetailListResponse is the response body when fetching a list of Transfer Details
type TransferDetailListResponse struct {
	Header *coreapi.ResponseHeader               `json:"header"`
	Data   []*transferdetails.TransferDetailData `json:"data"`
}

func toTransferDetailCreatePayload(request CreateTransferDetailRequest) (*transferdetails.CreateTransferDetailRequest, error) {
	payload := new(transferdetails.CreateTransferDetailRequest)
	payload.DocumentID = request.DocumentID
	payload.Data = request.Data

	return payload, nil
}

func toTransferDetailUpdatePayload(request UpdateTransferDetailRequest) (*transferdetails.UpdateTransferDetailRequest, error) {
	payload := new(transferdetails.UpdateTransferDetailRequest)
	payload.DocumentID = request.DocumentID
	payload.Data = request.Data
	payload.TransferID = request.TransferID

	return payload, nil
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
