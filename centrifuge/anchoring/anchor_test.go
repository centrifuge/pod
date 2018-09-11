// +build unit

package anchoring

import (
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestNewAnchorId(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		err   string
	}{
		{
			"smallerSlice",
			tools.RandomSlice(AnchorIdLength - 1),
			"invalid length byte slice provided for anchorId",
		},
		{
			"largerSlice",
			tools.RandomSlice(AnchorIdLength + 1),
			"invalid length byte slice provided for anchorId",
		},
		{
			"nilSlice",
			nil,
			"invalid length byte slice provided for anchorId",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewAnchorId(test.slice)
			assert.Equal(t, test.err, err.Error())
		})
	}
}

func TestNewDocRoot(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		err   string
	}{
		{
			"smallerSlice",
			tools.RandomSlice(RootLength - 1),
			"invalid length byte slice provided for docRoot",
		},
		{
			"largerSlice",
			tools.RandomSlice(RootLength + 1),
			"invalid length byte slice provided for docRoot",
		},
		{
			"nilSlice",
			nil,
			"invalid length byte slice provided for docRoot",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewDocRoot(test.slice)
			assert.Equal(t, test.err, err.Error())
		})
	}
}
