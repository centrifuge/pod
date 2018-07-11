// +build unit

package tools_test

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestStringToByte32_andBack(t *testing.T) {
	//invalid input params
	bytes, err := tools.StringToByte32("too short")
	assert.EqualValuesf(t, [32]byte{}, bytes, "Should receive empty byte array if string is not 32 chars")
	assert.Error(t, err, "Should return error on invalid input parameter")

	bytes, err = tools.StringToByte32("")
	assert.EqualValuesf(t, [32]byte{}, bytes, "Should receive empty byte array if string is not 32 chars")
	assert.Error(t, err, "Should return error on invalid input parameter")

	bytes, err = tools.StringToByte32("too long. 12345678901234567890123456789032")
	assert.EqualValuesf(t, [32]byte{}, bytes, "Should receive empty byte array if string is not 32 chars")
	assert.Error(t, err, "Should return error on invalid input parameter")

	//valid input param
	convertThis := "12345678901234567890123456789032"
	bytes, err = tools.StringToByte32(convertThis)
	assert.Nil(t, err, "Should not return error on 32 length string")

	assert.EqualValues(t, []byte("12345678901234567890123456789032")[:], bytes[:])

	convertedBack, _ := tools.Byte32ToString(bytes)
	assert.EqualValues(t, convertThis, convertedBack, "Converted back value should be the same as original input")
}

func TestRandomByte32(t *testing.T) {
	random := tools.RandomByte32()
	assert.NotNil(t, random, "Should receive non-nil")
	assert.NotEqual(t, [32]byte{}, random, "Should receive a filled byte array")
}

func TestRandomString32(t *testing.T) {
	random := tools.RandomString32()
	assert.NotNil(t, random, "Should receive non-nil")
	assert.NotEqual(t, "", random, "Should receive a filled string")
	assert.Equal(t, 32, len(random), "Should receive 32 long string")
}

func TestIsEmptyByte(t *testing.T) {
	assert.True(t, tools.IsEmptyByteSlice([]byte{}))
	assert.False(t, tools.IsEmptyByteSlice([]byte{'1', '1'}))
}

func TestIsEmptyByte32(t *testing.T) {
	assert.True(t, tools.IsEmptyByte32([32]byte{}))
	assert.False(t, tools.IsEmptyByte32([32]byte{'1', '1'}))
}

var testDataIsSameByteSlice = []struct {
	a        []byte
	b        []byte
	expected bool
}{
	{
		[]byte("abc1"),
		[]byte("abc1"),
		true,
	},
	{
		nil,
		nil,
		true,
	},
	{
		[]byte{},
		[]byte{},
		true,
	},
	{
		nil,
		[]byte("abc1"),
		false,
	},
	{
		[]byte("abc1"),
		nil,
		false,
	},
	{
		[]byte("abc1"),
		[]byte("abc2"),
		false,
	},
}

func TestIsSameByteSlice(t *testing.T) {
	for _, tt := range testDataIsSameByteSlice {
		actual := tools.IsSameByteSlice(tt.a, tt.b)
		assert.Equal(t, tt.expected, actual)
	}
}

func TestByte32toByte(t *testing.T) {
	b32, err := tools.StringToByte32("12345678901234567890123456789032")
	assert.Nil(t, err)

	actual := tools.Byte32ToByteArray(b32)
	exp := []byte("12345678901234567890123456789032")
	assert.Truef(t, tools.IsSameByteSlice(exp, actual), "Expected to be [%v] but got [%v]", exp, actual)

	actual = tools.Byte32ToByteArray([32]byte{})
	exp = []byte{}
	assert.Truef(t, tools.IsSameByteSlice(exp, actual), "Expected to be [%v] but got [%v]", exp, actual)
}

func TestByteArrayToByte32(t *testing.T){
	exp := [32]byte{}
	act := [32]byte{}
	tst := []byte{}

	tst = []byte("12345678901234567890123456789032")
	exp, err := tools.StringToByte32("12345678901234567890123456789032")
	assert.Nil(t, err)
	act, err = tools.ByteArrayToByte32(tst)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)

	tst = []byte{}
	exp = [32]byte{}
	act, err = tools.ByteArrayToByte32(tst)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)

	tst = []byte("123456789012345678901234567890321")
	exp = [32]byte{}
	act, err = tools.ByteArrayToByte32(tst)
	assert.Error(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)
}