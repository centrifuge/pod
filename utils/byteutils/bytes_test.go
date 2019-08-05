// +build unit

package byteutils

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestAddZeroBytesSuffix(t *testing.T) {
	tests := []struct {
		data     []byte
		required int
		result   []byte
	}{

		// nil data
		{},

		// 0 required with data
		{
			data:   []byte{1, 2, 3},
			result: []byte{1, 2, 3},
		},

		// data length > required
		{
			data:     []byte{1, 2, 3},
			required: 2,
			result:   []byte{1, 2, 3},
		},

		// data length == required
		{
			data:     []byte{1, 2, 3},
			required: 3,
			result:   []byte{1, 2, 3},
		},

		// data length < require
		{
			data:     []byte{1, 2, 3},
			required: 4,
			result:   []byte{1, 2, 3, 0},
		},

		// nil data with required > 0
		{
			required: 4,
			result:   []byte{0, 0, 0, 0},
		},
	}

	for _, c := range tests {
		res := AddZeroBytesSuffix(c.data, c.required)
		assert.Equal(t, c.result, res)
	}
}

func TestRemoveZeroBytesSuffix(t *testing.T) {
	tests := []struct {
		data, result []byte
	}{
		// nil data
		{},

		// data with no suffix zeroes
		{
			data:   []byte{1, 2, 3},
			result: []byte{1, 2, 3},
		},

		// data with suffix zeros
		{
			data:   []byte{1, 2, 3, 0, 0},
			result: []byte{1, 2, 3},
		},

		// data with alternate zeroes
		{
			data:   []byte{1, 2, 3, 0, 1, 0},
			result: []byte{1, 2, 3, 0, 1},
		},
	}

	for _, c := range tests {
		res := RemoveZeroBytesSuffix(c.data)
		assert.Equal(t, c.result, res)
	}
}

func TestIntBytesFromString(t *testing.T) {
	// empty
	// zero
	// invalid
	// actual
	tests := []struct {
		s     string
		res   []byte
		error bool
	}{
		// empty
		{},

		// zero
		{
			s:   "000000",
			res: []byte{},
		},

		// invalid
		{
			s:     "invalid",
			error: true,
		},

		// success
		{
			s:   "99999999999999999999",
			res: []byte{0x5, 0x6b, 0xc7, 0x5e, 0x2d, 0x63, 0xf, 0xff, 0xff},
		},
	}

	for _, c := range tests {
		res, err := IntBytesFromString(c.s)
		if c.error {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, c.res, res)
	}
}

func TestIsBitSet(t *testing.T) {
	tests := []struct {
		name   string
		b      byte
		p      uint
		expect bool
	}{
		{
			"all 1s, pos 0",
			byte(255),
			0,
			true,
		},
		{
			"all 1s, pos 1",
			byte(255),
			1,
			true,
		},
		{
			"all 1s, pos 7",
			byte(255),
			7,
			true,
		},
		{
			"pos[1] = 1, pos 1",
			byte(2),
			1,
			true,
		},
		{
			"pos[1] = 1, pos 2",
			byte(2),
			2,
			false,
		},
		{
			"all 0s, pos 0",
			byte(0),
			0,
			false,
		},
		{
			"all 0s, pos 1",
			byte(0),
			1,
			false,
		},
		{
			"all 0s, pos 7",
			byte(0),
			7,
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, IsBitSet(test.b, test.p))
		})
	}
}

func TestSetBit(t *testing.T) {
	tests := []struct {
		name   string
		b      byte
		p      uint
		expect byte
	}{
		{
			"val = 1, set pos = 1",
			byte(1),
			1,
			3,
		},
		{
			"val = 1, set pos = 2",
			byte(1),
			2,
			5,
		},
		{
			"val = 2, set pos = 2",
			byte(2),
			2,
			6,
		},
		{
			"val = 254, set pos = 0",
			byte(254),
			0,
			255,
		},
		{
			"val = 254, set pos = 8",
			byte(254),
			9,
			254,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, SetBit(test.b, test.p))
		})
	}
}

func TestClearBit(t *testing.T) {
	tests := []struct {
		name   string
		b      byte
		p      uint
		expect byte
	}{
		{
			"val = 1, clear pos = 1",
			byte(1),
			1,
			1,
		},
		{
			"val = 1, clear pos = 0",
			byte(1),
			0,
			0,
		},
		{
			"val = 2, clear pos = 1",
			byte(2),
			1,
			0,
		},
		{
			"val = 255, clear pos = 0",
			byte(255),
			0,
			254,
		},
		{
			"val = 254, clear pos = 8",
			byte(255),
			9,
			255,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, ClearBit(test.b, test.p))
		})
	}
}

func TestSortByte32Slice(t *testing.T) {
	bts := [][32]byte{
		utils.RandomByte32(),
		utils.RandomByte32(),
		utils.RandomByte32(),
		utils.RandomByte32(),
	}

	sortBytes := SortByte32Slice(bts)
	// pre element must be less than equal to next one
	for i := 1; i < len(sortBytes); i++ {
		assert.NotEqual(t, bytes.Compare(sortBytes[i-1][:], sortBytes[i][:]), 1)
	}
}

func TestHexBytes_MarshalJSON(t *testing.T) {
	tests := []struct {
		d []byte
	}{
		// nil
		{},

		{
			d: []byte{},
		},

		{
			d: utils.RandomSlice(10),
		},
	}

	type testHex struct {
		Hex HexBytes `json:"hex"`
	}

	for _, c := range tests {
		h := HexBytes(c.d)
		th := testHex{Hex: h}

		d, err := json.Marshal(th)
		assert.NoError(t, err)

		nth := new(testHex)
		err = json.Unmarshal(d, nth)
		assert.NoError(t, err)
		assert.Equal(t, len(c.d), len(nth.Hex.Bytes()))
		if len(c.d) < 1 {
			continue
		}

		assert.Equal(t, c.d, nth.Hex.Bytes())
	}
}

func TestHexBytes_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		str string
		d   []byte
		err bool
	}{
		// nil
		{str: "{}"},

		// empty
		{str: `{"hex": ""}`, err: true},

		// invalid
		{str: `{"hex": "12345dswr"}`, err: true},

		// valid
		{str: `{"hex": "0x"}`},

		// valid
		{str: `{"hex": "0xde0a3a82af"}`, d: []byte{222, 10, 58, 130, 175}},
	}

	type testHex struct {
		Hex HexBytes `json:"hex"`
	}

	for _, c := range tests {
		th := new(testHex)
		err := json.Unmarshal([]byte(c.str), th)
		if c.err {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, len(c.d), len(th.Hex.Bytes()))
		if len(c.d) < 1 {
			continue
		}

		assert.Equal(t, c.d, th.Hex.Bytes())
		assert.Equal(t, hexutil.Encode(c.d), th.Hex.String())
	}
}

func TestOptionalHex_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		str string
		d   []byte
		err bool
	}{
		// nil
		{str: "{}"},

		// empty
		{str: `{"hex": ""}`},

		// invalid
		{str: `{"hex": "12345dswr"}`, err: true},

		// valid
		{str: `{"hex": "0x"}`},

		// valid
		{str: `{"hex": "0xde0a3a82af"}`, d: []byte{222, 10, 58, 130, 175}},
	}

	type testHex struct {
		Hex OptionalHex `json:"hex"`
	}

	for _, c := range tests {
		th := new(testHex)
		err := json.Unmarshal([]byte(c.str), th)
		if c.err {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, len(c.d), len(th.Hex.Bytes()))
		if len(c.d) < 1 {
			continue
		}

		assert.Equal(t, c.d, th.Hex.Bytes())
		assert.Equal(t, hexutil.Encode(c.d), th.Hex.String())
	}
}
