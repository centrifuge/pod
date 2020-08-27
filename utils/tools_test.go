// +build unit

package utils

import (
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	var tst []byte

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

func TestMustSliceToByte32(t *testing.T) {
	// 32 bytes
	eout := RandomByte32()
	in := make([]byte, 32, 32)
	copy(in, eout[:])

	out := MustSliceToByte32(in)
	assert.Equal(t, eout, out)

	// less than 32 bytes
	in = RandomSlice(30)
	eout = [32]byte{}
	copy(eout[:], in)
	out = MustSliceToByte32(in)
	assert.Equal(t, eout, out)

	// more than 32(panic case)
	in = RandomSlice(34)
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()
	MustSliceToByte32(in)
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

func TestSliceOfByteSlicesToHexStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input [][]byte
	}{
		{
			name:  "happy",
			input: [][]byte{RandomSlice(32)},
		},
		{
			name:  "empty",
			input: [][]byte{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := SliceOfByteSlicesToHexStringSlice(test.input)
			for _, s := range out {
				verifyHex(t, s)
			}
		})
	}
}

func TestConvertIntToBytes(t *testing.T) {
	n := 5
	nb, err := ConvertIntToByte32(n)
	assert.NoError(t, err)
	ni := ConvertByte32ToInt(nb)
	assert.Equal(t, n, ni)
}

func TestAddressTo32Bytes(t *testing.T) {
	address := RandomSlice(common.AddressLength)
	address32bytes := AddressTo32Bytes(common.BytesToAddress(address))
	for i := 0; i < common.AddressLength; i++ {

		assert.Equal(t, address[i], address32bytes[i+32-common.AddressLength], "every byte should be equal")
	}
	for i := 0; i < 32-common.AddressLength; i++ {
		assert.Equal(t, uint8(0x0), address32bytes[i], "first 12 bytes need to be equal 0")
	}
}

func verifyHex(t *testing.T, val string) {
	_, err := hexutil.Decode(val)
	assert.Nil(t, err)
}

func TestRandomBigInt(t *testing.T) {
	tests := []struct {
		max   string
		isErr bool
	}{
		{
			"999",
			false,
		},
		{
			"150",
			false,
		},
		{
			"999999999999999",
			false,
		},
		{
			"10000",
			false,
		},
		{
			"323hu",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.max, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				n, err := RandomBigInt(test.max)
				if test.isErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					tt := new(big.Int)
					tt.SetString(test.max, 10)
					assert.True(t, n.Cmp(tt) <= 0)
				}
			}
		})
	}
}

func TestInRange(t *testing.T) {
	for i := 0; i < 10; i++ {
		n := InRange(i, 0, 10)
		assert.True(t, n)
	}
	n := 5
	assert.False(t, InRange(n, 6, 10))
}
