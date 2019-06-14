package transferdetails

import (
	"reflect"

	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Data is the default transfer details extension schema
type Data struct {
	TransferId           string `json:"transfer_id,omitempty" attr:"bytes"`
	SenderId			  string `json:"sender_id,omitempty" attr:"bytes"`
	RecipientId			  string `json:"recipient_id,omitempty" attr:"bytes"`
	ScheduledDate      string `json:"scheduled_date,omitempty" attr:"timestamp"`
	SettlementDate string `json:"settlement_date,omitempty" attr:"timestamp"`
	SettlementReference string `json:"settlement_reference,omitempty" attr:"timestamp"`
	Amount       string `json:"amount,omitempty" attr:"decimal"`
	// the currency and amount will be combined once we have standardised multiformats
	Currency              string `json:"currency,omitempty" attr:"string"`
	Status string `json:"status,omitempty" attr:"string"`
	TransferType string `json:"transfer_type,omitempty" attr:"string"`
	Data string `json:"data,omitempty" attr:"string"`
	NftAddress            string `json:"nft_address,omitempty" attr:"bytes"`
}

//make generic?
func (f *Data) initTransferDetailsFromData(data *clientfunpb.FundingData) {
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
