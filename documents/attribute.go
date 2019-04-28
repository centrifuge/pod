package documents

import (
	"math/big"
	"reflect"

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
		// TODO IMPORTANT!!! use our own type for int256 with size checks
		return reflect.TypeOf(&big.Int{}), nil
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

// Attribute represents a custom attribute of a document
type Attribute struct {
	AttrType attributeType `json:"attr_type"`
	KeyLabel string        `json:"key_label"`
	Key      AttrKey       `json:"key"`
	Value    interface{}   `json:"value"`
}

// newAttribute creates a new custom attribute
func newAttribute(keyLabel string, attributeType attributeType, value interface{}) (Attribute, error) {
	if keyLabel == "" {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with an empty string as name"))
	}

	if value == nil {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with a nil value"))
	}

	tp, err := allowedAttributeTypes(attributeType)
	if err != nil {
		return Attribute{}, err
	}

	if !reflect.TypeOf(value).AssignableTo(tp) {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, errors.New("provided type doesn't match the actual type of the value"))
	}

	hashedKey, err := NewAttrKey(keyLabel)
	if err != nil {
		return Attribute{}, errors.NewTypedError(ErrCDAttribute, err)
	}

	return Attribute{
		KeyLabel: keyLabel,
		Key:      hashedKey,
		AttrType: attributeType,
		Value:    value,
	}, nil
}

func (a *Attribute) copy() Attribute {
	return Attribute{a.AttrType, a.KeyLabel, a.Key, a.Value}
}
