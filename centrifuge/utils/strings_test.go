// +build unit

package utils

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestStringLengthEqual(t *testing.T) {
	tests := []struct {
		msg    string
		len    int
		result bool
	}{
		{
			"data",
			4,
			true,
		},

		{
			"dataa",
			4,
			false,
		},

		{
			"",
			4,
			false,
		},
	}

	for _, c := range tests {
		got := StringLengthEqual(c.msg, c.len)
		assert.Equal(t, c.result, got, "result must match")
	}
}

func TestStringEmpty(t *testing.T) {
	tests := []struct {
		msg    string
		result bool
	}{
		{
			result: true,
		},

		{
			"data",
			false,
		},
	}

	for _, c := range tests {
		got := StringEmpty(c.msg)
		assert.Equal(t, c.result, got, "result must match")
	}
}
