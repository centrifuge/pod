// +build unit

package coredocument

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoreDocumentProcessor_SendNilDocument(t *testing.T) {
	err := GetDefaultCoreDocumentProcessor().Send(nil, nil, "")

	assert.Error(t, err, "should have thrown an error")
}

func TestCoreDocumentProcessor_AnchorNilDocument(t *testing.T) {
	err := GetDefaultCoreDocumentProcessor().Anchor(nil)

	assert.Error(t, err, "should have thrown an error")
}

func TestCheck32BytesFilled(t *testing.T) {
	valid := make([]byte, 32)
	valid[0] = 0x1
	assert.True(t, Check32BytesFilled(valid))

	invalid := make([]byte, 32)
	assert.False(t, Check32BytesFilled(invalid))

	for i := 0; i < 32; i++ {
		invalid[i] = 0x0
	}
	assert.False(t, Check32BytesFilled(invalid))
}
