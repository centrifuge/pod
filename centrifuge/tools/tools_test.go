// +build unit

package tools

import (
	"testing"

	"encoding/binary"

	"github.com/stretchr/testify/assert"
)

func TestRandomByte32(t *testing.T) {
	random := RandomByte32()
	assert.NotNil(t, random, "Should receive non-nil")
	assert.NotEqual(t, [32]byte{}, random, "Should receive a filled byte array")
}

func TestRandomSliceN(t *testing.T) {
	random := RandomSlice(32)
	assert.NotNil(t, random, "should receive non-nil")
	assert.False(t, IsEmptyByteSlice(random))
	assert.Len(t, random, 32)
}

func TestIsEmptyByte(t *testing.T) {
	assert.True(t, IsEmptyByteSlice([]byte{}))
	assert.False(t, IsEmptyByteSlice([]byte{'1', '1'}))
}

func TestIsEmptyByte32(t *testing.T) {
	assert.True(t, IsEmptyByte32([32]byte{}))
	assert.False(t, IsEmptyByte32([32]byte{'1', '1'}))
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
		actual := IsSameByteSlice(tt.a, tt.b)
		assert.Equal(t, tt.expected, actual)
	}
}

func TestCheckMultiple32BytesFilled(t *testing.T) {
	emptySlice := make([]byte, 32)
	filledSlice := make([]byte, 32)
	filledSlice2 := make([]byte, 32)

	for i := 0; i < 32; i++ {
		filledSlice[i] = 0x1
		filledSlice2[i] = 0x1
	}
	assert.True(t, CheckMultiple32BytesFilled(filledSlice))
	assert.True(t, CheckMultiple32BytesFilled(filledSlice, filledSlice2))
	assert.False(t, CheckMultiple32BytesFilled(emptySlice, filledSlice))
	assert.False(t, CheckMultiple32BytesFilled(filledSlice, emptySlice))
}

func TestByte32ToSlice(t *testing.T) {
	b32Empty := [32]byte{}
	assert.Equal(t, []byte{}, Byte32ToSlice(b32Empty))

	b32Full := [32]byte{}
	expectedSlice := make([]byte, 32)
	for i := 0; i < 32; i++ {
		b32Full[i] = 0x1
		expectedSlice[i] = 0x1
	}
	assert.Equal(t, expectedSlice, Byte32ToSlice(b32Full))

}

func TestSliceToByte32(t *testing.T) {
	exp := [32]byte{}
	act := [32]byte{}
	tst := []byte{}

	tst = []byte("12345678901234567890123456789032")
	copy(exp[:], tst[:32])
	act, err := SliceToByte32(tst)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)

	tst = []byte{}
	exp = [32]byte{}
	act, err = SliceToByte32(tst)
	assert.Nil(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)

	tst = []byte("123456789012345678901234567890321")
	exp = [32]byte{}
	act, err = SliceToByte32(tst)
	assert.Error(t, err)
	assert.EqualValues(t, exp, act, "Expected to be [%v] but got [%v]", exp, act)
}

func TestByteSliceToBigInt(t *testing.T) {
	// uint32
	expected := uint32(15)
	byteVal := make([]byte, 4)
	binary.BigEndian.PutUint32(byteVal, expected)
	bigInt := ByteSliceToBigInt(byteVal)
	actual := uint32(bigInt.Uint64())
	assert.Equal(t, expected, actual)

	// uint48
	tst := []byte{1, 2, 3, 4, 5, 6}
	bigInt = ByteSliceToBigInt(tst)
	assert.Equal(t, tst, bigInt.Bytes())
}

func TestByteFixedToBigInt(t *testing.T) {
	// uint32
	expected := uint32(15)
	byteVal := make([]byte, 4)
	binary.BigEndian.PutUint32(byteVal, expected)
	bigInt := ByteFixedToBigInt(byteVal, 4)
	actual := uint32(bigInt.Uint64())
	assert.Equal(t, expected, actual)

	// uint48
	tst := []byte{1, 2, 3, 4, 5, 6}
	bigInt = ByteFixedToBigInt(tst, 6)
	assert.Equal(t, tst, bigInt.Bytes())
}

func TestIsValidByteSliceForLength(t *testing.T) {
	tests := []struct {
		name   string
		slice  []byte
		length int
		result bool
	}{
		{
			"validByteSlice",
			RandomSlice(3),
			3,
			true,
		},
		{
			"smallerSlice",
			RandomSlice(2),
			3,
			false,
		},
		{
			"largerSlice",
			RandomSlice(4),
			3,
			false,
		},
		{
			"nilSlice",
			nil,
			3,
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, IsValidByteSliceForLength(test.slice, test.length))
		})
	}
}
