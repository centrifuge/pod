package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewAttribute(t *testing.T) {
	testdecimal := new(Decimal)
	err := testdecimal.SetString("5.1321312")
	assert.NoError(t, err)
	tests := []struct {
		name        string
		readableKey string
		attrType    AllowedAttributeType
		value       interface{}
		at          *attribute
		errs        bool
		errStr      string
	}{
		{
			"readable key empty",
			"",
			Str,
			"",
			nil,
			true,
			"can't create attribute with an empty string as name",
		},
		{
			"value nil",
			"somekey",
			Str,
			nil,
			nil,
			true,
			"can't create attribute with a nil value",
		},
		{
			"type mismatch",
			"somekey",
			Str,
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
			Str,
			"someval",
			&attribute{
				attrType:    Str,
				readableKey: "string",
				value:       "someval",
			},
			false,
			"",
		},
		//{
		//	"int256",
		//	"int256",
		//	Int256,
		//	testdecimal,
		//	&attribute{
		//		attrType:    Int256,
		//		readableKey: "int256",
		//		value:       big.NewInt(123),
		//	},
		//	false,
		//	"",
		//},
		{
			"bigdecimal",
			"bigdecimal",
			BigDec,
			testdecimal,
			&attribute{
				attrType:    BigDec,
				readableKey: "bigdecimal",
				value:       testdecimal,
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
				assert.Equal(t, attr.hashedKey, hashedKey)
				assert.Equal(t, attr.attrType, test.at.attrType)
				assert.Equal(t, attr.value, test.at.value)
				assert.Equal(t, attr.readableKey, test.at.readableKey)
			}
		})
	}
}
