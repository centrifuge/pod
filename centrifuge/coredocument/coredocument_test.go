// +build unit

package coredocument

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoreDocumentProcessor_SendNilDocument(t *testing.T) {
	err := GetDefaultCoreDocumentProcessor().Send(nil, nil, []byte{})

	assert.Error(t, err, "should have thrown an error")
}

func TestCoreDocumentProcessor_AnchorNilDocument(t *testing.T) {
	err := GetDefaultCoreDocumentProcessor().Anchor(nil)

	assert.Error(t, err, "should have thrown an error")
}
