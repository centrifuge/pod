package documents

import (
	"reflect"
	"strconv"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// attributeType represents the custom attribute types allowed in models
type attributeType string

// String string repr
func (a attributeType) String() string {
	return string(a)
}

const (
	// Int256Type is the standard integer custom attribute type
	Int256Type attributeType = "int256"

	// BigDecType is the standard big decimal custom attribute type
	BigDecType attributeType = "bigdecimal"

	// StrType is the standard string custom attribute type
	StrType attributeType = "string"

	// BytsType is the standard bytes custom attribute type
	BytsType attributeType = "bytes"

	// TimestmpType is the standard time stamp custom attribute type
	TimestmpType attributeType = "timestamp"
)

func allowedAttributeTypes(typ attributeType) (reflect.Type, error) {
	switch typ {
	case Int256Type:
		return reflect.TypeOf(&Int256{}), nil
	case BigDecType:
		return reflect.TypeOf(&Decimal{}), nil
	case StrType:
		return reflect.TypeOf(""), nil
	case BytsType:
		return reflect.TypeOf([]byte{}), nil
	case TimestmpType:
		return reflect.TypeOf(int64(1)), nil
	default:
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't find the given attribute in allowed attribute types"))
	}
}

// AttrKey represents a sha256 hash of a attribute label given by a user.
type AttrKey [32]byte

// NewAttrKey creates a new AttrKey from label
func NewAttrKey(label string) (AttrKey, error) {
	hashedKey, err := crypto.Sha256Hash([]byte(label))
	if err != nil {
		return AttrKey{}, errors.NewTypedError(ErrCDAttribute, err)
	}
	var a [32]byte
	copy(a[:], hashedKey)
	return AttrKey(a), nil
}

// AttrKeyFromBytes converts bytes to AttrKey
func AttrKeyFromBytes(b []byte) AttrKey {
	var a [32]byte
	copy(a[:], b)
	return a
}

// String converts the AttrKey to a hex string
func (a AttrKey) String() string {
	return hexutil.Encode(a[:])
}

// MarshalText converts the AttrKey to its text form
func (a AttrKey) MarshalText() (text []byte, err error) {
	return []byte(hexutil.Encode(a[:])), nil
}

// UnmarshalText converts text to AttrKey
func (a *AttrKey) UnmarshalText(text []byte) error {
	b, err := hexutil.Decode(string(text))
	if err != nil {
		return err
	}
	*a = AttrKeyFromBytes(b)
	return nil
}

// AttrVal represents a strongly typed value of an attribute
type AttrVal struct {
	AttrType attributeType
	I256Val  *Int256
	DecVal   *Decimal
	StrVal   string
	BytVal   []byte
	TSVal    int64
}

func NewAttrVal(attributeType attributeType, value interface{}) (AttrVal, error) {
	tp, err := allowedAttributeTypes(attributeType)
	if err != nil {
		return AttrVal{}, err
	}

	if !reflect.TypeOf(value).AssignableTo(tp) {
		return AttrVal{}, errors.NewTypedError(ErrCDAttribute, errors.New("provided type doesn't match the actual type of the value"))
	}

	a := AttrVal{AttrType: attributeType}
	switch attributeType {
	case Int256Type:
		a.I256Val = value.(*Int256)
	case BigDecType:
		a.DecVal = value.(*Decimal)
	case StrType:
		a.StrVal = value.(string)
	case BytsType:
		a.BytVal = value.([]byte)
	case TimestmpType:
		a.TSVal = value.(int64)
	default:
		return AttrVal{}, errors.NewTypedError(ErrCDAttribute, errors.New("can't find the given attribute in allowed attribute types"))
	}
	return a, nil
}

// Attribute represents a custom attribute of a document
type Attribute struct {
	KeyLabel string  `json:"key_label"`
	Key      AttrKey `json:"key"`
	Value    AttrVal `json:"value"`
}

// newAttribute creates a new custom attribute
func newAttribute(keyLabel string, attributeType attributeType, value interface{}) (Attribute, error) {
	if keyLabel == "" {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with an empty string as name"))
	}

	if value == nil {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with a nil value"))
	}

	attrVal, err := NewAttrVal(attributeType, value)
	if err != nil {
		return Attribute{}, err
	}

	hashedKey, err := NewAttrKey(keyLabel)
	if err != nil {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, err)
	}

	return Attribute{
		KeyLabel: keyLabel,
		Key:      hashedKey,
		Value:    attrVal,
	}, nil
}

// strToAttrVal converts a string value of an attribute to its proper type
func strToAttrVal(typ attributeType, value string) (interface{}, error) {
	switch typ {
	case Int256Type:
		return NewInt256(value)
	case BigDecType:
		return NewDecimal(value)
	case StrType:
		return value, nil
	case BytsType:
		return hexutil.Decode(value)
	case TimestmpType:
		return strconv.Atoi(value)
	default:
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("given value is not a valid attribute type"))
	}
}

// attrValToStr converts an attribute value to its string repr
func attrValToStr(value AttrVal) string {
	switch value.AttrType {
	case Int256Type:
		return value.I256Val.String()
	case BigDecType:
		return value.DecVal.String()
	case StrType:
		return value.StrVal
	case BytsType:
		return hexutil.Encode(value.BytVal)
	case TimestmpType:
		return string(value.TSVal)
	default:
		log.Error("value: %v seems to be corrupt", value)
		return ""
	}
}

func (a *Attribute) copy() Attribute {
	return Attribute{a.KeyLabel, a.Key, a.Value}
}
