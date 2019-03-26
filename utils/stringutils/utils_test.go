// +build unit

package stringutils

import (
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
