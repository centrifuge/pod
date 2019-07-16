package funding

import (
	"reflect"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// TODO: generic? Create something like ServiceRegistry for Custom Attributes...
func (f *OldData) initFundingFromData(data *funpb.FundingData) {
	types := reflect.TypeOf(*f)
	values := reflect.ValueOf(*data)
	for i := 0; i < types.NumField(); i++ {
		n := types.Field(i).Name
		v := values.FieldByName(n).Interface().(string)
		// converter assumes string struct fields
		reflect.ValueOf(f).Elem().FieldByName(n).SetString(v)

	}
}

// TODO: generic?
func (f *OldData) getClientData() *funpb.FundingData {
	clientData := new(funpb.FundingData)
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
