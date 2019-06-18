package transferdetails

//func (f *TransferDetailData) initTransferFromData(data *TransferDetailData) {
//	types := reflect.TypeOf(*f)
//	values := reflect.ValueOf(*data)
//	for i := 0; i < types.NumField(); i++ {
//		n := types.Field(i).Name
//		v := values.FieldByName(n).Interface().(string)
//		// converter assumes string struct fields
//		reflect.ValueOf(f).Elem().FieldByName(n).SetString(v)
//
//	}
//}
//
//func (f *TransferDetailData) getClientData() *TransferDetailData {
//	clientData := new(TransferDetailData)
//	types := reflect.TypeOf(*f)
//	values := reflect.ValueOf(*f)
//	for i := 0; i < types.NumField(); i++ {
//		n := types.Field(i).Name
//		v := values.FieldByName(n).Interface().(string)
//		// converter assumes string struct fields
//		reflect.ValueOf(clientData).Elem().FieldByName(n).SetString(v)
//
//	}
//	return clientData
//}
