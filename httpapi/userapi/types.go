// nolint
package userapi

import (
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
)

// TODO: think: generic custom attribute set creation?

//CreateTransferDetailRequest is the request body for creating a Transfer Detail
type CreateTransferDetailRequest struct {
	DocumentID string                        `json:"document_id"`
	Data       extensions.TransferDetailData `json:"data"`
}

// UpdateTransferDetailRequest is the request body for updating a Transfer Detail
type UpdateTransferDetailRequest struct {
	DocumentID string                        `json:"document_id"`
	TransferID string                        `json:"transfer_id"`
	Data       extensions.TransferDetailData `json:"data"`
}

// TransferDetailResponse is the response body when fetching a Transfer Detail
type TransferDetailResponse struct {
	Header coreapi.ResponseHeader        `json:"header"`
	Data   extensions.TransferDetailData `json:"data"`
}

// TransferDetailListResponse is the response body when fetching a list of Transfer Details
type TransferDetailListResponse struct {
	Header coreapi.ResponseHeader          `json:"header"`
	Data   []extensions.TransferDetailData `json:"data"`
}

func toTransferDetailCreatePayload(request CreateTransferDetailRequest) (*extensions.CreateTransferDetailRequest, error) {
	payload := extensions.CreateTransferDetailRequest{
		DocumentID: request.DocumentID,
		Data:       request.Data,
	}
	return &payload, nil
}

func toTransferDetailUpdatePayload(request UpdateTransferDetailRequest) (*extensions.UpdateTransferDetailRequest, error) {
	payload := extensions.UpdateTransferDetailRequest{
		DocumentID: request.DocumentID,
		Data:       request.Data,
		TransferID: request.TransferID,
	}
	return &payload, nil
}
