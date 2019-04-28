package documents

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

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
		at          *Attribute
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
			&Attribute{
				KeyLabel: "string",
				Value: AttrVal{
					StrVal:   "someval",
					AttrType: StrType,
				},
			},
			false,
			"",
		},
		{
			"int256",
			"int256",
			Int256Type,
			&Int256{*big.NewInt(123)},
			&Attribute{
				KeyLabel: "int256",
				Value: AttrVal{
					I256Val:  &Int256{*big.NewInt(123)},
					AttrType: Int256Type,
				},
			},
			false,
			"",
		},
		{
			"bigdecimal",
			"bigdecimal",
			BigDecType,
			testdecimal,
			&Attribute{
				KeyLabel: "bigdecimal",
				Value: AttrVal{
					DecVal:   testdecimal,
					AttrType: BigDecType,
				},
			},
			false,
			"",
		},
		{
			"bytes",
			"bytes",
			BytsType,
			[]byte{1},
			&Attribute{
				KeyLabel: "bytes",
				Value: AttrVal{
					BytVal:   []byte{1},
					AttrType: BytsType,
				},
			},
			false,
			"",
		},
		{
			"timestamp",
			"timestamp",
			TimestmpType,
			ttime.Unix(),
			&Attribute{
				KeyLabel: "timestamp",
				Value: AttrVal{
					TSVal:    ttime.Unix(),
					AttrType: TimestmpType,
				},
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
				hashedKey, err := NewAttrKey(test.at.KeyLabel)
				assert.NoError(t, err)
				if assert.NotNil(t, attr) {
					assert.Equal(t, hashedKey, attr.Key)
					assert.Equal(t, test.at.Value.AttrType, attr.Value.AttrType)
					assert.Equal(t, test.at.Value, attr.Value)
					assert.Equal(t, test.at.KeyLabel, attr.KeyLabel)
				}
			}
		})
	}
}

func TestAttrKey_MarshalText(t *testing.T) {
	a, err := NewAttrKey("somekey")
	assert.NoError(t, err)
	m := map[AttrKey]string{a: "dwefw"}
	mstr, err := json.Marshal(m)
	assert.NoError(t, err)
	m1 := make(map[AttrKey]string)
	err = json.Unmarshal(mstr, &m1)
	assert.NoError(t, err)
	assert.Equal(t, m[a], m1[a])
}
