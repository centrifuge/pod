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
}
