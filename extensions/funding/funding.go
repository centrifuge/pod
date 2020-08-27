package funding

// Data is the default funding extension schema.
type Data struct {
	AgreementID           string `json:"agreement_id,omitempty" attr:"bytes"`
	BorrowerID            string `json:"borrower_id,omitempty" attr:"bytes"`
	FunderID              string `json:"funder_id,omitempty" attr:"bytes"`
	Amount                string `json:"amount,omitempty" attr:"decimal"`
	Apr                   string `json:"apr,omitempty" attr:"string"`
	Days                  string `json:"days,omitempty" attr:"integer"`
	Fee                   string `json:"fee,omitempty" attr:"decimal"`
	RepaymentDueDate      string `json:"repayment_due_date,omitempty" attr:"timestamp"`
	RepaymentOccurredDate string `json:"repayment_occurred_date,omitempty" attr:"timestamp"`
	RepaymentAmount       string `json:"repayment_amount,omitempty" attr:"decimal"`
	Currency              string `json:"currency,omitempty" attr:"string"`
	NFTAddress            string `json:"nft_address,omitempty" attr:"bytes"`
	PaymentDetailsID      string `json:"payment_details_id,omitempty" attr:"bytes"`
}

// Signature is the funding agreement Signature.
type Signature struct {
	Valid             string `json:"valid"`
	OutdatedSignature string `json:"outdated_signature"`
	Identity          string `json:"identity"`
	SignedVersion     string `json:"signed_version"`
}
