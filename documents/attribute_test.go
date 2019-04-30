package documents

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestAttribute_isAttrTypeAllowed(t *testing.T) {
	tests := []struct {
		attrType attributeType
		result   bool
	}{
		{
			attrType: AttrDecimal,
			result:   true,
		},

		{
			attrType: attributeType("some type"),
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
		attrType    attributeType
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
			attr, err := newAttribute(test.readableKey, test.attrType, test.value)
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
	a, err := AttrKeyFromLabel("somekey")
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
		tp    attributeType
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
			"unknown type",
			attributeType("some type"),
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

			v.Type = attributeType("some type")
			_, err = v.String()
			assert.Error(t, err)
		})
	}
}
