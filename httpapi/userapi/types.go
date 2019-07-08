// nolint
package userapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common"
)

// TODO: think: generic custom attribute set creation?

//CreateTransferDetailRequest is the request body for creating a Transfer Detail
type CreateTransferDetailRequest struct {
	DocumentID string               `json:"document_id"`
	Data       transferdetails.Data `json:"data"`
}

// UpdateTransferDetailRequest is the request body for updating a Transfer Detail
type UpdateTransferDetailRequest struct {
	DocumentID string               `json:"document_id"`
	TransferID string               `json:"transfer_id"`
	Data       transferdetails.Data `json:"data"`
}

// TransferDetailResponse is the response body when fetching a Transfer Detail
type TransferDetailResponse struct {
	Header coreapi.ResponseHeader `json:"header"`
	Data   transferdetails.Data   `json:"data"`
}

// TransferDetailListResponse is the response body when fetching a list of Transfer Details
type TransferDetailListResponse struct {
	Header coreapi.ResponseHeader `json:"header"`
	Data   []transferdetails.Data `json:"data"`
}

func toTransferDetailCreatePayload(request CreateTransferDetailRequest) (*transferdetails.CreateTransferDetailRequest, error) {
	payload := transferdetails.CreateTransferDetailRequest{
		DocumentID: request.DocumentID,
		Data:       request.Data,
	}
	return &payload, nil
}

func toTransferDetailUpdatePayload(request UpdateTransferDetailRequest) (*transferdetails.UpdateTransferDetailRequest, error) {
	payload := transferdetails.UpdateTransferDetailRequest{
		DocumentID: request.DocumentID,
		Data:       request.Data,
		TransferID: request.TransferID,
	}
	return &payload, nil
}

// CreateInvoiceRequest defines the payload for creating documents.
type CreateInvoiceRequest struct {
	ReadAccess  []common.Address     `json:"read_access" swaggertype:"array,string"`
	WriteAccess []common.Address     `json:"write_access" swaggertype:"array,string"`
	Data        invoice.Data         `json:"data"`
	Attributes  coreapi.AttributeMapRequest `json:"attributes"`
}

// CreatePurchaseOrderRequest holds details for creating Purchase order Document.
type CreatePurchaseOrderRequest struct {
	ReadAccess  []identity.DID              `json:"read_access" swaggertype:"array,string"`
	WriteAccess []identity.DID              `json:"write_access" swaggertype:"array,string"`
	Data        purchaseorder.Data          `json:"data"`
	Attributes  coreapi.AttributeMapRequest `json:"attributes"`
}

// PurchaseOrderResponse represents the purchase order in client API format.
type PurchaseOrderResponse struct {
	Header     coreapi.ResponseHeader       `json:"header"`
	Data       purchaseorder.Data           `json:"data"`
	Attributes coreapi.AttributeMapResponse `json:"attributes"`
}

func toPurchaseOrderResponse(model documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp PurchaseOrderResponse, err error) {
	docResp, err := coreapi.GetDocumentResponse(model, tokenRegistry, jobID)
	if err != nil {
		return resp, err
	}

	return PurchaseOrderResponse{
		Header:     docResp.Header,
		Attributes: docResp.Attributes,
		Data:       docResp.Data.(purchaseorder.Data),
	}, nil
}
