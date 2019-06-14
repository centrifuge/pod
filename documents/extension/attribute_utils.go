package extension

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/extension/funding"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"reflect"
	"strings"
)

const (
	fundingLabel              = "funding_agreement"
	fundingFieldKey           = "funding_agreement[{IDX}]."
	idxKey                    = "{IDX}"
)

func newAgreementID() string {
	return hexutil.Encode(utils.RandomSlice(32))
}

func generateLabel(field, idx, fieldName string) string {
	return strings.Replace(field, idxKey, idx, -1) + fieldName
}

func labelFromJSONTag(idx, jsonTag string, fieldKey string) string {
	jsonKeyIdx := 0
	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag, ",")
	return generateLabel(fieldKey, idx, jsonParts[jsonKeyIdx])
}

func getArrayLatestIDX(model documents.Model, arrayLabel string) (idx *documents.Int256, err error) {
	key, err := documents.AttrKeyFromLabel(arrayLabel)
	if err != nil {
		return nil, err
	}

	attr, err := model.GetAttribute(key)
	if err != nil {
		return nil, err
	}

	idx = attr.Value.Int256

	z, err := documents.NewInt256("0")
	if err != nil {
		return nil, err
	}

	// idx < 0
	if idx.Cmp(z) == -1 {
		return nil, funding.ErrFundingIndex
	}

	return idx, nil

}

func incrementArrayAttrIDX(model documents.Model, arrayLabel string) (attr documents.Attribute, err error) {
	key, err := documents.AttrKeyFromLabel(arrayLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewAttribute(arrayLabel, documents.AttrInt256, "0")
	}

	idx, err := getArrayLatestIDX(model, arrayLabel)
	if err != nil {
		return attr, err
	}

	// increment idx
	newIdx, err := idx.Inc()

	if err != nil {
		return attr, err
	}

	return documents.NewAttribute(arrayLabel, documents.AttrInt256, newIdx.String())
}

func fillAttributeList(data interface{}, idx string, fieldKey string) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	types := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < types.NumField(); i++ {

		value := values.Field(i).Interface().(string)
		if value != "" {
			jsonKey := types.Field(i).Tag.Get("json")
			label := labelFromJSONTag(idx, jsonKey, fieldKey)

			attrType := types.Field(i).Tag.Get("attr")
			attr, err := documents.NewAttribute(label, documents.AttributeType(attrType), value)
			if err != nil {
				return nil, err
			}

			attributes = append(attributes, attr)
		}

	}
	return attributes, nil
}

func createAttributesList(current documents.Model, data interface{}, fieldKey string) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := incrementArrayAttrIDX(current, fundingLabel)
	if err != nil {
		return nil, err
	}

	attributes, err = fillAttributeList(data, idx.Value.Int256.String(), fieldKey)
	if err != nil {
		return nil, err
	}

	// add updated idx
	attributes = append(attributes, idx)

	return attributes, nil

}
