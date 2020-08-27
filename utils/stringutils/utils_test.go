// +build unit

package stringutils

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		strs, res []string
	}{
		// no strings
		{},

		// no duplicates
		{
			strs: []string{"a", "B", "c"},
			res:  []string{"a", "B", "c"},
		},

		// duplicates
		{
			strs: []string{"a", "B", "c", "b", "C", "D"},
			res:  []string{"a", "B", "c", "D"},
		},
	}

	for _, c := range tests {
		res := RemoveDuplicates(c.strs)
		assert.Equal(t, c.res, res)
	}
}

func TestContainsStringMatch(t *testing.T) {
	m := "some.something\\[.*\\].any"
	str := "some.something[0x1234567890].any"
	assert.True(t, ContainsStringMatch(m, str))

	m = "nothing"
	assert.False(t, ContainsStringMatch(m, str))
}

func TestContainsStringMatchInSlice(t *testing.T) {
	m := []string{"some.something\\[.*\\].any", "blabla"}
	str := "some.something[0x1234567890].any"
	assert.True(t, ContainsStringMatchInSlice(m, str))

	m = []string{"nothing", "blabla"}
	assert.False(t, ContainsStringMatchInSlice(m, str))
}

func TestContainsBytesMatch(t *testing.T) {
	m0 := []byte{3, 0, 0, 0, 0, 0, 0, 1}
	m1 := []byte{0, 0, 0, 4}
	m := fmt.Sprintf("%s(.{32})%s", hex.EncodeToString(m0), hex.EncodeToString(m1))
	val := append(m0, append([]byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 254, 5}, m1...)...)
	assert.True(t, ContainsBytesMatch(m, val))

	m1 = []byte{0, 0, 0, 3}
	val = append(m0, append([]byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 254, 5}, m1...)...)
	assert.False(t, ContainsBytesMatch(m, val))
}
