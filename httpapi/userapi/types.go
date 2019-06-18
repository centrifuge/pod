// nolint
package userapi

import (
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
)

// TransferDetailData is the default transfer details extension schema
type TransferDetailData struct {
	TransferId          string `json:"transfer_id,omitempty" attr:"bytes"`
	SenderId            string `json:"sender_id,omitempty" attr:"bytes"`
	RecipientId         string `json:"recipient_id,omitempty" attr:"bytes"`
	ScheduledDate       string `json:"scheduled_date,omitempty" attr:"timestamp"`
	SettlementDate      string `json:"settlement_date,omitempty" attr:"timestamp"`
	SettlementReference string `json:"settlement_reference,omitempty" attr:"timestamp"`
	Amount              string `json:"amount,omitempty" attr:"decimal"`
	// the currency and amount will be combined once we have standardised multiformats
	Currency     string `json:"currency,omitempty" attr:"string"`
	Status       string `json:"status,omitempty" attr:"string"`
	TransferType string `json:"transfer_type,omitempty" attr:"string"`
	Data         string `json:"data,omitempty" attr:"string"`
}

// TODO: think: generic custom attribute set creation?

//CreateTransferDetailRequest
type CreateTransferDetailRequest struct {
	Identifier string                              `json:"identifier" swaggertype:"primitive,string"`
	Data       *transferdetails.TransferDetailData `json:"data"`
}

// UpdateTransferDetailRequest
type UpdateTransferDetailRequest struct {
	Identifier string                              `json:"identifier" swaggertype:"primitive,string"`
	TransferId string                              `json:"transfer_id" swaggertype:"primitive,string"`
	Data       *transferdetails.TransferDetailData `json:"data"`
}

// TransferDetailResponse
type TransferDetailResponse struct {
	Header *documentpb.ResponseHeader          `json:"header"`
	Data   *transferdetails.TransferDetailData `json:"data"`
}

// TransferDetailListResponse
type TransferDetailListResponse struct {
	Header *documentpb.ResponseHeader            `json:"header"`
	Data   []*transferdetails.TransferDetailData `json:"data"`
}

//GetRequest
//GetVersionRequest
