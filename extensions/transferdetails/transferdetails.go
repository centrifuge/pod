package transferdetails

// TransferDetailData is the default transfer details extension schema
type TransferDetailData struct {
	TransferID          string `json:"transfer_id,omitempty" attr:"bytes"`
	SenderID            string `json:"sender_id,omitempty" attr:"bytes"`
	RecipientID         string `json:"recipient_id,omitempty" attr:"bytes"`
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
	DocumentID string
	Data       *TransferDetailData
}

// UpdateTransferDetailRequest holds the required fields to update a transfer agreement
type UpdateTransferDetailRequest struct {
	DocumentID string
	TransferID string
	Data       *TransferDetailData
}

// TransferDetail holds a TransferDetail response
type TransferDetail struct {
	Data *TransferDetailData
}

// TransferDetailList holds a list of TransferDetails
type TransferDetailList struct {
	Data []*TransferDetailData
}
