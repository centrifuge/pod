//go:build unit

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
		got := IsStringOfLength(c.msg, c.len)
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
		got := IsStringEmpty(c.msg)
		assert.Equal(t, c.result, got, "result must match")
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		str    string
		slice  []string
		result bool
	}{
		// empty everything
		{},
		// empty string
		{
			slice: []string{"abc"},
		},

		// empty slice
		{
			str: "abc",
		},

		// missing str
		{
			str:   "test",
			slice: []string{"abc", "bcd", "cde"},
		},

		// success
		{
			str:    "abc",
			slice:  []string{"bcd", "cde", "abc"},
			result: true,
		},
	}

	for _, c := range tests {
		if ok := ContainsString(c.slice, c.str); ok != c.result {
			t.Fatalf("ContainsString(%v, %s)=%v; expected %v", c.slice, c.str, c.result, ok)
		}
	}
}
