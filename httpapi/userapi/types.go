package userapi

// TransferDetailData is the default transfer details extension schema
type TransferDetailData struct {
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
}



////make generic?
//func (t *TransferDetailData) initTransferDetailsFromData(data *TransferDetailData) {
//	types := reflect.TypeOf(*t)
//	values := reflect.ValueOf(*data)
//	for i := 0; i < types.NumField(); i++ {
//		n := types.Field(i).Name
//		v := values.FieldByName(n).Interface().(string)
//		// converter assumes string struct fields
//		reflect.ValueOf(*t).Elem().FieldByName(n).SetString(v)
//
//	}
//}
//
//func (t *TransferDetailData) getClientData() *TransferDetailData {
//	transferData := new(TransferDetailData)
//	types := reflect.TypeOf(*t)
//	values := reflect.ValueOf(*t)
//	for i := 0; i < types.NumField(); i++ {
//		n := types.Field(i).Name
//		v := values.FieldByName(n).Interface().(string)
//		// converter assumes string struct fields
//		reflect.ValueOf(transferData).Elem().FieldByName(n).SetString(v)
//
//	}
//	return transferData
//}