package documents

import (
	"fmt"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AttributeType represents the custom attribute type.
type AttributeType string

// String returns the readable name of the attribute type.
func (a AttributeType) String() string {
	return string(a)
}

// MonetaryType represents the monetary type of the attribute
type MonetaryType string

// String returns the readable name of the monetary type.
func (a MonetaryType) String() string {
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

	// AttrSigned is the custom signature attribute type
	AttrSigned AttributeType = "signed"

	// AttrMonetary is the monetary attribute type
	AttrMonetary AttributeType = "monetary"

	// MonetaryToken is the monetary type for tokens
	MonetaryToken MonetaryType = "token"
)

// isAttrTypeAllowed checks if the given attribute type is implemented and returns its `reflect.Type` if allowed.
func isAttrTypeAllowed(attr AttributeType) bool {
	switch attr {
	case AttrInt256, AttrDecimal, AttrString, AttrBytes, AttrTimestamp, AttrSigned, AttrMonetary:
		return true
	default:
		return false
	}
}

// AttrKey represents a sha256 hash of a attribute label given by a user.
type AttrKey [32]byte

// AttrKeyFromLabel creates a new AttrKey from label.
func AttrKeyFromLabel(label string) (attrKey AttrKey, err error) {
	if strings.TrimSpace(label) == "" {
		return attrKey, ErrEmptyAttrLabel
	}

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

// Signed is a custom attribute type with signature.
type Signed struct {
	Identity                                     *types.AccountID
	Type                                         AttributeType
	DocumentVersion, Value, Signature, PublicKey []byte
}

// String returns the hex value of the signature.
func (s Signed) String() string {
	return s.Identity.ToHexString()
}

// Monetary is a custom attribute type for monetary values
type Monetary struct {
	Value   *Decimal
	ChainID []byte
	Type    MonetaryType
	ID      []byte // Currency USD|0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2(DAI)|ETH
}

// String returns the readable representation of the monetary value
func (m Monetary) String() string {
	chStr := ""
	if len(m.ChainID) > 0 {
		chStr = "@" + hexutil.Encode(m.ChainID)
	}
	mID := string(m.ID)
	if m.Type == MonetaryToken {
		mID = hexutil.Encode(m.ID)
	}
	return fmt.Sprintf("%s %s%s", m.Value.String(), mID, chStr)
}

// AttrVal represents a strongly typed value of an attribute
type AttrVal struct {
	Type      AttributeType
	Int256    *Int256
	Decimal   *Decimal
	Str       string
	Bytes     []byte
	Timestamp *timestamppb.Timestamp
	Signed    Signed
	Monetary  Monetary
}

// ToBytes encodes attribute value into bytes.
func (attrVal AttrVal) ToBytes() ([]byte, error) {
	switch attrVal.Type {
	case AttrInt256:
		b := attrVal.Int256.Bytes()
		return b[:], nil
	case AttrDecimal:
		return attrVal.Decimal.Bytes()
	case AttrString:
		return []byte(attrVal.Str), nil
	case AttrBytes:
		return attrVal.Bytes, nil
	case AttrTimestamp:
		return byteutils.TimestampToBytes(attrVal.Timestamp, maxTimeByteLength)
	default:
		return nil, ErrNotValidAttrType
	}
}

// AttrValFromString converts the string value to necessary type based on the attribute type.
func AttrValFromString(attrType AttributeType, value string) (attrVal AttrVal, err error) {
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
		t, err = time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return attrVal, err
		}
		attrVal.Timestamp = timestamppb.New(t.UTC())
	default:
		return attrVal, ErrNotValidAttrType
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
		if !attrVal.Timestamp.IsValid() {
			return str, ErrInvalidAttrTimestamp
		}
		str = attrVal.Timestamp.AsTime().UTC().Format(time.RFC3339Nano)
	case AttrSigned:
		str = attrVal.Signed.String()
	case AttrMonetary:
		str = attrVal.Monetary.String()
	}

	return str, err
}

// Attribute represents a custom attribute of a document
type Attribute struct {
	KeyLabel string
	Key      AttrKey
	Value    AttrVal
}

// NewStringAttribute creates a new custom attribute.
func NewStringAttribute(keyLabel string, attrType AttributeType, value string) (attr Attribute, err error) {
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

// NewMonetaryAttribute creates new instance of Monetary Attribute
func NewMonetaryAttribute(keyLabel string, value *Decimal, chainID []byte, id string) (attr Attribute, err error) {
	if value == nil {
		return attr, errors.NewTypedError(ErrWrongAttrFormat, errors.New("empty value field"))
	}

	attrKey, err := AttrKeyFromLabel(keyLabel)
	if err != nil {
		return attr, err
	}

	token := MonetaryToken
	idb, err := hexutil.Decode(id)
	if err != nil {
		token = ""
		idb = []byte(id)
	}

	if len(idb) > monetaryIDLength {
		return attr, errors.NewTypedError(ErrWrongAttrFormat, errors.New("monetaryIDLength exceeds 32 bytes"))
	}

	attrVal := AttrVal{
		Type:     AttrMonetary,
		Monetary: Monetary{Value: value, Type: token, ChainID: chainID, ID: idb},
	}

	return Attribute{
		KeyLabel: keyLabel,
		Key:      attrKey,
		Value:    attrVal,
	}, nil
}

// NewSignedAttribute returns a new signed attribute
// takes keyLabel, signer identity, signer account, model and value
// doc version is next version of the document since that is the document version in which the attribute is added.
// signature payload: sign(identity + docID + docNextVersion + value)
// Note: versionID should always be the next version that is going to be anchored.
func NewSignedAttribute(keyLabel string, identity *types.AccountID, account config.Account, docID, versionID, value []byte, valType AttributeType) (attr Attribute, err error) {
	attrKey, err := AttrKeyFromLabel(keyLabel)
	if err != nil {
		return attr, err
	}

	signPayload := attributeSignaturePayload(identity.ToBytes(), docID, versionID, value)
	sig, err := account.SignMsg(signPayload)
	if err != nil {
		return attr, err
	}

	attrVal := AttrVal{
		Type: AttrSigned,
		Signed: Signed{
			Identity:        identity,
			DocumentVersion: versionID,
			Value:           value,
			Signature:       sig.Signature,
			PublicKey:       sig.PublicKey,
			Type:            valType,
		},
	}

	return Attribute{
		KeyLabel: keyLabel,
		Key:      attrKey,
		Value:    attrVal,
	}, nil
}

// attributeSignaturePayload creates the payload for signing an attribute
func attributeSignaturePayload(identity, id, version, value []byte) []byte {
	var signPayload []byte
	signPayload = append(signPayload, identity...)
	signPayload = append(signPayload, id...)
	signPayload = append(signPayload, version...)
	signPayload = append(signPayload, value...)
	return signPayload
}
