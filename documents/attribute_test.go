package documents

import (
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewAttribute(t *testing.T) {
	testdecimal := new(Decimal)
	err := testdecimal.SetString("5.1321312")
	assert.NoError(t, err)
	ttime := time.Now()
	tests := []struct {
		name        string
		readableKey string
		attrType    attributeType
		value       interface{}
		at          *attribute
		errs        bool
		errStr      string
	}{
		{
			"readable key empty",
			"",
			StrType,
			"",
			nil,
			true,
			"can't create attribute with an empty string as name",
		},
		{
			"value nil",
			"somekey",
			StrType,
			nil,
			nil,
			true,
			"can't create attribute with a nil value",
		},
		{
			"type mismatch",
			"somekey",
			StrType,
			12,
			nil,
			true,
			"provided type doesn't match the actual type of the value",
		},
		{
			"type not allowed",
			"somekey",
			"int",
			12,
			nil,
			true,
			"can't find the given attribute in allowed attribute types",
		},
		{
			"string",
			"string",
			StrType,
			"someval",
			&attribute{
				attrType:    StrType,
				readableKey: "string",
				value:       "someval",
			},
			false,
			"",
		},
		{
			"int256",
			"int256",
			Int256Type,
			big.NewInt(123),
			&attribute{
				attrType:    Int256Type,
				readableKey: "int256",
				value:       big.NewInt(123),
			},
			false,
			"",
		},
		{
			"bigdecimal",
			"bigdecimal",
			BigDecType,
			testdecimal,
			&attribute{
				attrType:    BigDecType,
				readableKey: "bigdecimal",
				value:       testdecimal,
			},
			false,
			"",
		},
		{
			"bytes",
			"bytes",
			BytsType,
			[]byte{1},
			&attribute{
				attrType:    BytsType,
				readableKey: "bytes",
				value:       []byte{1},
			},
			false,
			"",
		},
		{
			"timestamp",
			"timestamp",
			TimestmpType,
			ttime.Unix(),
			&attribute{
				attrType:    TimestmpType,
				readableKey: "timestamp",
				value:       ttime.Unix(),
			},
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attr, err := newAttribute(test.readableKey, test.attrType, test.value)
			if test.errs {
				if assert.Error(t, err) {
					assert.True(t, errors.IsOfType(ErrCDAttribute, err))
					assert.Contains(t, err.Error(), test.errStr)
				} else {
					t.Fail()
				}
			} else {
				assert.NoError(t, err)
				hashedKey, err := crypto.Sha256Hash([]byte(test.at.readableKey))
				assert.NoError(t, err)
				if assert.NotNil(t, attr) {
					assert.Equal(t, attr.hashedKey, hashedKey)
					assert.Equal(t, attr.attrType, test.at.attrType)
					assert.Equal(t, attr.value, test.at.value)
					assert.Equal(t, attr.readableKey, test.at.readableKey)
				}
			}
		})
	}
}
