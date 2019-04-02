package documents

import (
	"math/big"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
)

// AllowedAttributeType represents the custom attribute types allowed in models
type AllowedAttributeType string

// String string repr
func (a AllowedAttributeType) String() string {
	return string(a)
}

const (
	Int256   AllowedAttributeType = "int256"
	BigDec   AllowedAttributeType = "bigdecimal"
	Str      AllowedAttributeType = "string"
	Byts     AllowedAttributeType = "bytes"
	Timestmp AllowedAttributeType = "timestamp"
)

func allowedAttributeTypes(typ AllowedAttributeType) (reflect.Type, error) {
	switch typ {
	case Int256:
		// TODO use our own type for int256 with checks
		return reflect.TypeOf(&big.Int{}), nil
	case BigDec:
		return reflect.TypeOf(&Decimal{}), nil
	case Str:
		return reflect.TypeOf(""), nil
	case Byts:
		return reflect.TypeOf([]byte{}), nil
	case Timestmp:
		return reflect.TypeOf(&time.Time{}), nil
	default:
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't find the given attribute in allowed attribute types"))
	}
}

// attribute represents a custom attribute of a document
type attribute struct {
	attrType    AllowedAttributeType
	readableKey string
	hashedKey   []byte
	value       interface{}
}

// newAttribute creates a new custom attribute
func newAttribute(readableKey string, attributeType AllowedAttributeType, value interface{}) (*attribute, error) {
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
