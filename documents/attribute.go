package documents

import (
	"time"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// AttributeType represents the custom attribute type.
type AttributeType string

// String returns the readable name of the attribute type.
func (a AttributeType) String() string {
	return string(a)
}

const (
	// AttrInt256 is the standard integer custom attribute type
	AttrInt256 AttributeType = "integer"

	// AttrDecimal is the standard big decimal custom attribute type
	AttrDecimal AttributeType = "decimal"

	// AttrString is the standard string custom attribute type
	AttrString AttributeType = "string"

	// AttrBytes is the standard bytes custom attribute type
	AttrBytes AttributeType = "bytes"

	// AttrTimestamp is the standard time stamp custom attribute type
	AttrTimestamp AttributeType = "timestamp"
)

// allowedAttrTypes holds a map of allowed attribute types and their reflect.Type
var allowedAttrTypes = map[AttributeType]struct{}{
	AttrInt256:    {},
	AttrDecimal:   {},
	AttrString:    {},
	AttrBytes:     {},
	AttrTimestamp: {},
}

// isAttrTypeAllowed checks if the given attribute type is implemented and returns its `reflect.Type` if allowed.
func isAttrTypeAllowed(attr AttributeType) bool {
	_, ok := allowedAttrTypes[attr]
	return ok
}

// AttrKey represents a sha256 hash of a attribute label given by a user.
type AttrKey [32]byte

// AttrKeyFromLabel creates a new AttrKey from label.
func AttrKeyFromLabel(label string) (attrKey AttrKey, err error) {
	hashedKey, err := crypto.Sha256Hash([]byte(label))
	if err != nil {
		return attrKey, err
	}

	return AttrKeyFromBytes(hashedKey)
}

// AttrKeyFromBytes converts bytes to AttrKey
func AttrKeyFromBytes(b []byte) (AttrKey, error) {
	return utils.SliceToByte32(b)
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

	*a, err = AttrKeyFromBytes(b)
	return err
}

// AttrVal represents a strongly typed value of an attribute
type AttrVal struct {
	Type      AttributeType
	Int256    *Int256
	Decimal   *Decimal
	Str       string
	Bytes     []byte
	Timestamp *timestamp.Timestamp
}

// AttrValFromString converts the string value to necessary type based on the attribute type.
func AttrValFromString(attrType AttributeType, value string) (attrVal AttrVal, err error) {
	if !isAttrTypeAllowed(attrType) {
		return attrVal, ErrNotValidAttrType
	}

	attrVal.Type = attrType
	switch attrType {
	case AttrInt256:
		attrVal.Int256, err = NewInt256(value)
	case AttrDecimal:
		attrVal.Decimal, err = NewDecimal(value)
	case AttrString:
		attrVal.Str = value
	case AttrBytes:
		attrVal.Bytes, err = hexutil.Decode(value)
	case AttrTimestamp:
		var t time.Time
		t, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return attrVal, err
		}

		attrVal.Timestamp, err = utils.ToTimestamp(t)
	}

	return attrVal, err
}

// String returns the string representation of the AttrVal.
func (attrVal AttrVal) String() (str string, err error) {
	if !isAttrTypeAllowed(attrVal.Type) {
		return str, ErrNotValidAttrType
	}

	switch attrVal.Type {
	case AttrInt256:
		str = attrVal.Int256.String()
	case AttrDecimal:
		str = attrVal.Decimal.String()
	case AttrString:
		str = attrVal.Str
	case AttrBytes:
		str = hexutil.Encode(attrVal.Bytes)
	case AttrTimestamp:
		var tp time.Time
		tp, err = utils.FromTimestamp(attrVal.Timestamp)
		if err != nil {
			break
		}

		str = tp.Format(time.RFC3339)
	}

	return str, err
}

// Attribute represents a custom attribute of a document
type Attribute struct {
	KeyLabel string
	Key      AttrKey
	Value    AttrVal
}

// NewAttribute creates a new custom attribute.
func NewAttribute(keyLabel string, attrType AttributeType, value string) (attr Attribute, err error) {
	if keyLabel == "" {
		return attr, ErrEmptyAttrLabel
	}

	attrKey, err := AttrKeyFromLabel(keyLabel)
	if err != nil {
		return attr, err
	}

	attrVal, err := AttrValFromString(attrType, value)
	if err != nil {
		return attr, err
	}

	return Attribute{
		KeyLabel: keyLabel,
		Key:      attrKey,
		Value:    attrVal,
	}, nil
}
