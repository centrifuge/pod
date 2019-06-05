package funding

import (
	"reflect"

	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Data is the default funding extension schema
type Data struct {
	AgreementId           string `json:"agreement_id,omitempty" attr:"bytes"`
	BorrowerId			  string `json:"borrower_id,omitempty" attr:"bytes"`
	FunderId			  string `json:"funder_id,omitempty" attr:"bytes"`
	Amount                string `json:"amount,omitempty" attr:"decimal"`
	Apr                   string `json:"apr,omitempty" attr:"string"`
	Days                  string `json:"days,omitempty" attr:"integer"`
	Fee                   string `json:"fee,omitempty" attr:"decimal"`
	RepaymentDueDate      string `json:"repayment_due_date,omitempty" attr:"timestamp"`
	RepaymentOccurredDate string `json:"repayment_occurred_date,omitempty" attr:"timestamp"`
	RepaymentAmount       string `json:"repayment_amount,omitempty" attr:"decimal"`
	Currency              string `json:"currency,omitempty" attr:"string"`
	NftAddress            string `json:"nft_address,omitempty" attr:"bytes"`
	PaymentDetailsId      string `json:"payment_details_id,omitempty" attr:"bytes"`
}

func (f *Data) initFundingFromData(data *clientfunpb.FundingData) {
	types := reflect.TypeOf(*f)
	values := reflect.ValueOf(*data)
	for i := 0; i < types.NumField(); i++ {
		n := types.Field(i).Name
		v := values.FieldByName(n).Interface().(string)
		// converter assumes string struct fields
		reflect.ValueOf(f).Elem().FieldByName(n).SetString(v)

	}
}

func (f *Data) getClientData() *clientfunpb.FundingData {
	clientData := new(clientfunpb.FundingData)
	types := reflect.TypeOf(*f)
	values := reflect.ValueOf(*f)
	for i := 0; i < types.NumField(); i++ {
		n := types.Field(i).Name
		v := values.FieldByName(n).Interface().(string)
		// converter assumes string struct fields
		reflect.ValueOf(clientData).Elem().FieldByName(n).SetString(v)

	}
	return clientData
}
