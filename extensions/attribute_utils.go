package extensions

import (
	"reflect"
	"strings"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	idxKey = "{IDX}"
)

// NewAttributeSetID generates an identifier for a new attribute set
func NewAttributeSetID() string {
	return hexutil.Encode(utils.RandomSlice(32))
}

// GenerateLabel generates a field label
func GenerateLabel(field, idx, fieldName string) string {
	return strings.Replace(field, idxKey, idx, -1) + fieldName
}

// LabelFromJSONTag converts a JSON tag to a label
func LabelFromJSONTag(idx, jsonTag string, fieldKey string) string {
	jsonKeyIdx := 0
	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag, ",")
	return GenerateLabel(fieldKey, idx, jsonParts[jsonKeyIdx])
}

// GetArrayLatestIDX gets the last index from an array of attributes
func GetArrayLatestIDX(model documents.Model, typeLabel string) (idx *documents.Int256, err error) {
	key, err := documents.AttrKeyFromLabel(typeLabel)
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
		return nil, ErrArrayIndex
	}

	return idx, nil

}

// IncrementArrayAttrIDX increments the index of an array for a new attribute
func IncrementArrayAttrIDX(model documents.Model, typeLabel string) (attr documents.Attribute, err error) {
	key, err := documents.AttrKeyFromLabel(typeLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewStringAttribute(typeLabel, documents.AttrInt256, "0")
	}

	idx, err := GetArrayLatestIDX(model, typeLabel)
	if err != nil {
		return attr, err
	}

	// increment idx
	newIdx, err := idx.Inc()
	if err != nil {
		return attr, err
	}

	return documents.NewStringAttribute(typeLabel, documents.AttrInt256, newIdx.String())
}

// FillAttributeList fills an attributes list from the JSON object
func FillAttributeList(data interface{}, idx, fieldKey string) ([]documents.Attribute, error) {
	var attributes []documents.Attribute
	types := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < types.NumField(); i++ {
		value := values.Field(i).Interface().(string)
		if value != "" {
			jsonKey := types.Field(i).Tag.Get("json")
			label := LabelFromJSONTag(idx, jsonKey, fieldKey)

			attrType := types.Field(i).Tag.Get("attr")
			attr, err := documents.NewStringAttribute(label, documents.AttributeType(attrType), value)
			if err != nil {
				return nil, err
			}

			attributes = append(attributes, attr)
		}

	}
	return attributes, nil
}

// CreateAttributesList creates an attributes list on a passed in document
func CreateAttributesList(current documents.Model, data interface{}, fieldKey, typeLabel string) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := IncrementArrayAttrIDX(current, typeLabel)
	if err != nil {
		return nil, err
	}

	attributes, err = FillAttributeList(data, idx.Value.Int256.String(), fieldKey)
	if err != nil {
		return nil, err
	}

	// add updated idx
	attributes = append(attributes, idx)
	return attributes, nil
}

// DeleteAttributesSet deletes attributes that already exist on a given model for the addition of new attributes to the set
func DeleteAttributesSet(model documents.Model, data interface{}, idx, fieldKey string) (documents.Model, error) {
	types := reflect.TypeOf(data)
	for i := 0; i < types.NumField(); i++ {
		jsonKey := types.Field(i).Tag.Get("json")
		key, err := documents.AttrKeyFromLabel(LabelFromJSONTag(idx, jsonKey, fieldKey))
		if err != nil {
			continue
		}

		if model.AttributeExists(key) {
			err := model.DeleteAttribute(key, false)
			if err != nil {
				return nil, err
			}
		}

	}
	return model, nil
}

// FindAttributeSetIDX returns the index of an attribute set given given its descriptive label
func FindAttributeSetIDX(model documents.Model, attributeSetID, typeLabel, idLabel, fieldKey string) (idx string, err error) {
	lastIdx, err := GetArrayLatestIDX(model, typeLabel)
	if err != nil {
		return idx, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return idx, err
	}

	for i.Cmp(lastIdx) != 1 {
		label := GenerateLabel(fieldKey, i.String(), idLabel)
		k, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return idx, err
		}

		attr, err := model.GetAttribute(k)
		if err != nil {
			return idx, err
		}

		attrFundingID, err := attr.Value.String()
		if err != nil {
			return idx, err
		}
		if attrFundingID == attributeSetID {
			return i.String(), nil
		}
		i, err = i.Inc()

		if err != nil {
			return idx, err
		}

	}

	return idx, ErrAttributeSetNotFound
}

// ToMapAttributes converts an array of documents.Attributes to a map
func ToMapAttributes(attrs []documents.Attribute) map[documents.AttrKey]documents.Attribute {
	if len(attrs) < 1 {
		return nil
	}

	m := make(map[documents.AttrKey]documents.Attribute)
	for _, v := range attrs {
		m[v.Key] = documents.Attribute{
			KeyLabel: v.KeyLabel,
			Key:      v.Key,
			Value:    v.Value,
		}
	}

	return m
}

// TODO: placeholder below for generic finding and deriving data from attribute sets
