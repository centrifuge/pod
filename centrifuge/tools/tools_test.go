// +build unit

package tools_test

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomByte32(t *testing.T) {
	random := tools.RandomByte32()
	assert.NotNil(t, random, "Should receive non-nil")
	assert.NotEqual(t, [32]byte{}, random, "Should receive a filled byte array")
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

func TestSliceToByte32(t *testing.T) {
	exp := [32]byte{}
	act := [32]byte{}
	tst := []byte{}

	tst = []byte("12345678901234567890123456789032")
	copy(exp[:], tst[:32])
	act, err := tools.SliceToByte32(tst)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)

	tst = []byte{}
	exp = [32]byte{}
	act, err = tools.SliceToByte32(tst)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)

	tst = []byte("123456789012345678901234567890321")
	exp = [32]byte{}
	act, err = tools.SliceToByte32(tst)
	assert.Error(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)
}
