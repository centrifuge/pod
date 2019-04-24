package documents

import (
	"math/big"
	"reflect"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
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

// attribute represents a custom attribute of a document
type attribute struct {
	attrType    attributeType
	readableKey string
	hashedKey   []byte
	value       interface{}
}

// newAttribute creates a new custom attribute
func newAttribute(readableKey string, attributeType attributeType, value interface{}) (*attribute, error) {
	if readableKey == "" {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with an empty string as name"))
	}

	if value == nil {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with a nil value"))
	}

	tp, err := allowedAttributeTypes(attributeType)
	if err != nil {
		return nil, err
	}

	if !reflect.TypeOf(value).AssignableTo(tp) {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("provided type doesn't match the actual type of the value"))
	}

	hashedKey, err := crypto.Sha256Hash([]byte(readableKey))
	if err != nil {
		return nil, errors.NewTypedError(ErrCDAttribute, err)
	}

	return &attribute{
		readableKey: readableKey,
		hashedKey:   hashedKey,
		attrType:    attributeType,
		value:       value,
	}, nil
}

func (a *attribute) copy() *attribute {
	return &attribute{a.attrType, a.readableKey, a.hashedKey, a.value}
}
