// nolint
package transferdetails

import "github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"

// TransferDetailData is the default transfer details extension schema
type TransferDetailData struct {
	TransferId          string `json:"transfer_id,omitempty" attr:"bytes"`
	SenderId            string `json:"sender_id,omitempty" attr:"bytes"`
	RecipientId         string `json:"recipient_id,omitempty" attr:"bytes"`
	ScheduledDate       string `json:"scheduled_date,omitempty" attr:"timestamp"`
	SettlementDate      string `json:"settlement_date,omitempty" attr:"timestamp"`
	SettlementReference string `json:"settlement_reference,omitempty" attr:"bytes"`
	Amount              string `json:"amount,omitempty" attr:"decimal"`
	// the currency and amount will be combined once we have standardised multiformats
	Currency     string `json:"currency,omitempty" attr:"string"`
	Status       string `json:"status,omitempty" attr:"string"`
	TransferType string `json:"transfer_type,omitempty" attr:"string"`
	Data         string `json:"data,omitempty" attr:"bytes"`
}

// TODO: make these generic? CreateAttributeSetRequest?
// CreateTransferDetailRequest holds the required fields to create a new transfer agreement
type CreateTransferDetailRequest struct {
	Identifier string
	Data       *TransferDetailData
}

// UpdateTransferDetailRequest holds the required fields to update a transfer agreement
type UpdateTransferDetailRequest struct {
	Identifier string
	TransferId string
	Data       *TransferDetailData
}

// TransferDetailResponse
type TransferDetailResponse struct {
	Header *documentpb.ResponseHeader
	Data   *TransferDetailData
}

// TransferDetailListResponse
type TransferDetailListResponse struct {
	Header *documentpb.ResponseHeader
	Data   []*TransferDetailData
}
