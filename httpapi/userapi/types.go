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
	payload.Data = request.Data
	payload.DocumentID = request.DocumentID

	return payload, nil
}

func toTransferDetailUpdatePayload(request UpdateTransferDetailRequest) (*transferdetails.UpdateTransferDetailRequest, error) {
	payload := new(transferdetails.UpdateTransferDetailRequest)
	payload.Data = request.Data
	payload.DocumentID = request.DocumentID
	payload.TransferID = request.TransferID

	return payload, nil
}
