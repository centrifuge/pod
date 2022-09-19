//go:build unit

package documents

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/errors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestAttribute_isAttrTypeAllowed(t *testing.T) {
	tests := []struct {
		attrType AttributeType
		result   bool
	}{
		{
			attrType: AttrDecimal,
			result:   true,
		},

		{
			attrType: AttributeType("some type"),
			result:   false,
		},
	}

	for _, c := range tests {
		assert.Equal(t, c.result, isAttrTypeAllowed(c.attrType))
	}
}

func TestNewAttribute(t *testing.T) {
	tests := []struct {
		name        string
		readableKey string
		attrType    AttributeType
		value       string
		errs        bool
		errStr      string
	}{
		{
			"readable key empty",
			"",
			AttrString,
			"",
			true,
			ErrEmptyAttrLabel.Error(),
		},

		{
			"type not allowed",
			"somekey",
			"some type",
			"",
			true,
			ErrNotValidAttrType.Error(),
		},

		{
			"string",
			"string",
			AttrString,
			"someval",
			false,
			"",
		},
		{
			"int256",
			"int256",
			AttrInt256,
			"123",
			false,
			"",
		},
		{
			"bigdecimal",
			"bigdecimal",
			AttrDecimal,
			"5.1321312",
			false,
			"",
		},
		{
			"bytes",
			"bytes",
			AttrBytes,
			hexutil.Encode([]byte{1}),
			false,
			"",
		},
		{
			"timestamp",
			"timestamp",
			AttrTimestamp,
			time.Now().UTC().Format(time.RFC3339),
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attr, err := NewStringAttribute(test.readableKey, test.attrType, test.value)
			if test.errs {
				assert.Error(t, err)
				assert.Equal(t, test.errStr, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.readableKey, attr.KeyLabel)
			assert.Equal(t, test.attrType, attr.Value.Type)
			attrKey, err := AttrKeyFromLabel(test.readableKey)
			assert.NoError(t, err)
			assert.Equal(t, attrKey, attr.Key)
			str, err := attr.Value.String()
			assert.NoError(t, err)
			assert.Equal(t, test.value, str)
		})
	}
}

func TestAttrKey(t *testing.T) {
	a, err := AttrKeyFromLabel("")
	assert.Error(t, err)
	a, err = AttrKeyFromLabel("somekey")
	assert.NoError(t, err)
	m := map[AttrKey]string{a: "dwefw"}
	mstr, err := json.Marshal(m)
	assert.NoError(t, err)
	m1 := make(map[AttrKey]string)
	err = json.Unmarshal(mstr, &m1)
	assert.NoError(t, err)
	assert.Equal(t, m[a], m1[a])
}

func TestAttrValFromString(t *testing.T) {
	tests := []struct {
		name  string
		tp    AttributeType
		value string
		error bool
	}{
		{
			"Int256",
			AttrInt256,
			"12343",
			false,
		},
		{
			"Decimal",
			AttrDecimal,
			"12343.2121",
			false,
		},
		{
			"string",
			AttrString,
			"123ewqewqer",
			false,
		},
		{
			"byte",
			AttrBytes,
			"0x12321abc",
			false,
		},
		{
			"timestamp",
			AttrTimestamp,
			time.Now().UTC().Format(time.RFC3339),
			false,
		},
		{
			"timestamp_nano",
			AttrTimestamp,
			time.Now().UTC().Format(time.RFC3339Nano),
			false,
		},
		{
			"unknown type",
			AttributeType("some type"),
			"",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := AttrValFromString(test.tp, test.value)
			if test.error {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			str, err := v.String()
			assert.NoError(t, err)
			assert.Equal(t, test.value, str)

			v.Type = AttributeType("some type")
			_, err = v.String()
			assert.Error(t, err)
		})
	}
}

func TestNewSignedAttribute(t *testing.T) {
	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// empty label
	_, err = NewSignedAttribute("", identity, nil, nil, nil, nil, AttrBytes)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEmptyAttrLabel, err))

	// failed sign
	label := "signed_label"

	id := utils.RandomSlice(32)
	version := utils.RandomSlice(32)
	value := utils.RandomSlice(50)

	epayload := attributeSignaturePayload(identity.ToBytes(), id, version, value)

	signErr := errors.New("error")

	acc := config.NewAccountMock(t)
	acc.On("SignMsg", epayload).
		Once().
		Return(nil, signErr)

	_, err = NewSignedAttribute(label, identity, acc, id, version, value, AttrBytes)
	assert.ErrorIs(t, err, signErr)

	// success
	signature := utils.RandomSlice(32)

	acc = config.NewAccountMock(t)
	acc.On("SignMsg", epayload).
		Once().
		Return(&coredocumentpb.Signature{Signature: signature}, nil)

	attr, err := NewSignedAttribute(label, identity, acc, id, version, value, AttrBytes)
	assert.NoError(t, err)

	attrKey, err := AttrKeyFromLabel(label)
	assert.NoError(t, err)
	assert.Equal(t, attrKey, attr.Key)
	assert.Equal(t, label, attr.KeyLabel)
	assert.Equal(t, AttrSigned, attr.Value.Type)
	assert.Equal(t, signature, attr.Value.Signed.Signature)
}

func TestNewMonetaryAttribute(t *testing.T) {
	dec, err := NewDecimal("1001.1001")
	assert.NoError(t, err)

	// empty label
	_, err = NewMonetaryAttribute("", dec, nil, "")
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEmptyAttrLabel, err))

	// monetary ID exceeded length
	label := "invoice_amount"
	chainID := []byte{1}
	idd := "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a29f8f72aa9304c8b593d555f12ef6589cc3a579a2" // 40 bytes
	_, err = NewMonetaryAttribute(label, dec, chainID, idd)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrWrongAttrFormat, err))

	// success fiat
	idd = "USD"
	attr, err := NewMonetaryAttribute(label, dec, chainID, idd)
	assert.NoError(t, err)
	assert.Equal(t, AttrMonetary, attr.Value.Type)
	attrKey, err := AttrKeyFromLabel(label)
	assert.NoError(t, err)
	assert.Equal(t, attrKey, attr.Key)
	assert.Equal(t, []byte(idd), attr.Value.Monetary.ID)
	assert.Equal(t, chainID, attr.Value.Monetary.ChainID)
	assert.Equal(t, "", attr.Value.Monetary.Type.String())
	assert.Equal(t, fmt.Sprintf("%s %s@%s", dec.String(), idd, hexutil.Encode(chainID)), attr.Value.Monetary.String())

	// success erc20
	idd = "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2"
	attr, err = NewMonetaryAttribute(label, dec, chainID, idd)
	assert.NoError(t, err)
	assert.Equal(t, AttrMonetary, attr.Value.Type)
	attrKey, err = AttrKeyFromLabel(label)
	assert.NoError(t, err)
	assert.Equal(t, attrKey, attr.Key)
	decIdd, err := hexutil.Decode(idd)
	assert.NoError(t, err)
	assert.Equal(t, decIdd, attr.Value.Monetary.ID)
	assert.Equal(t, chainID, attr.Value.Monetary.ChainID)
	assert.Equal(t, MonetaryToken, attr.Value.Monetary.Type)
	assert.Equal(t, fmt.Sprintf("%s %s@%s", dec.String(), idd, hexutil.Encode(chainID)), attr.Value.Monetary.String())
}

func TestGenerateDocumentSignatureProofField(t *testing.T) {
	// change with name of new keys in resources folder
	pub := "build/resources/signingKey.pub.pem"
	pvt := "build/resources/signingKey.key.pem"

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	pk, _, err := ed25519.GetSigningKeyPair(pub, pvt)
	assert.NoError(t, err)

	signerId := hexutil.Encode(append(identity.ToBytes(), pk...))
	signatureSender := fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerId)
	fmt.Println("SignatureSender", signatureSender)
}
