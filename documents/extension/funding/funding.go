package funding

import (
	"reflect"

	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Data is the default funding extension schema
type Data struct {
	FundingId             string `json:"funding_id,omitempty" attr:"string"`
	Amount                string `json:"amount,omitempty" attr:"string"`
	Apr                   string `json:"apr,omitempty" attr:"string"`
	Days                  string `json:"days,omitempty" attr:"string"`
	Fee                   string `json:"fee,omitempty" attr:"string"`
	RepaymentDueDate      string `json:"repayment_due_date,omitempty" attr:"string"`
	RepaymentOccurredDate string `json:"repayment_occurred_date,omitempty" attr:"string"`
	RepaymentAmount       string `json:"repayment_amount,omitempty" attr:"string"`
	Currency              string `json:"currency,omitempty" attr:"string"`
	NftAddress            string `json:"nft_address,omitempty" attr:"string"`
	PaymentDetailsId      string `json:"payment_details_id,omitempty" attr:"string"`
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
