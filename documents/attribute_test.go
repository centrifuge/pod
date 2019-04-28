package documents

import (
	"encoding/json"
	"math/big"
	"reflect"
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
			DecimalType,
			testdecimal,
			&Attribute{
				KeyLabel: "bigdecimal",
				Value: AttrVal{
					DecVal:   testdecimal,
					AttrType: DecimalType,
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

func TestStrToAttrVal(t *testing.T) {
	tests := []struct {
		name   string
		tp     attributeType
		value  string
		valTyp reflect.Type
	}{
		{
			"Int256",
			Int256Type,
			"12343",
			reflect.TypeOf(&Int256{}),
		},
		{
			"Decimal",
			DecimalType,
			"12343.2121",
			reflect.TypeOf(&Decimal{}),
		},
		{
			"string",
			StrType,
			"123ewqewqer",
			reflect.TypeOf("blah"),
		},
		{
			"byte",
			BytsType,
			"0x12321abc",
			reflect.TypeOf([]byte{}),
		},
		{
			"timestamp",
			TimestmpType,
			"1231231243",
			reflect.TypeOf(int(1)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := strToAttrVal(test.tp, test.value)
			assert.NoError(t, err)
			assert.True(t, reflect.TypeOf(v).AssignableTo(test.valTyp))
		})
	}
}

func TestAttrValToStr(t *testing.T) {
	testdecimal := new(Decimal)
	err := testdecimal.SetString("5.1321312")
	assert.NoError(t, err)
	tests := []struct {
		name   string
		value  AttrVal
		valStr string
	}{
		{
			"Int256",
			AttrVal{AttrType: Int256Type, I256Val: &Int256{*big.NewInt(-123)}},
			"-123",
		},
		{
			"Decimal",
			AttrVal{AttrType: DecimalType, DecVal: testdecimal},
			"5.1321312",
		},
		{
			"string",
			AttrVal{AttrType: StrType, StrVal: "123rewrew"},
			"123rewrew",
		},
		{
			"byte",
			AttrVal{AttrType: BytsType, BytVal: []byte{1, 2, 3}},
			"0x010203",
		},
		{
			"timestamp",
			AttrVal{AttrType: TimestmpType, TSVal: int64(2131)},
			"2131",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := attrValToStr(test.value)
			assert.Equal(t, test.valStr, v)
		})
	}
}
