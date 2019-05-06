package funding

import (
	"reflect"

	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Data is the default funding extension schema
type Data struct {
	FundingId             string `json:"funding_id,omitempty" attr:"bytes"`
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

func (f *Data) initFundingFromData(data *clientfundingpb.FundingData) {
	types := reflect.TypeOf(*f)
	values := reflect.ValueOf(*data)
	for i := 0; i < types.NumField(); i++ {
		n := types.Field(i).Name
		v := values.FieldByName(n).Interface().(string)
		reflect.ValueOf(f).Elem().FieldByName(n).SetString(v)

	}
}
