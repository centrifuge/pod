// +build unit

package byteutils

import (
	"testing"

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
