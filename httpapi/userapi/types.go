package userapi

// TransferDetailData is the default transfer details extension schema
type TransferDetailData struct {
	TransferID          string `json:"transfer_id,omitempty" attr:"bytes"`
	SenderID            string `json:"sender_id,omitempty" attr:"bytes"`
	RecipientID         string `json:"recipient_id,omitempty" attr:"bytes"`
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
