// +build unit

package coredocumentprocessor

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/stretchr/testify/assert"
)

// TODO(ved): more tests required for processor
var cdp defaultProcessor

func TestMain(m *testing.M) {
	cdp = defaultProcessor{}
	result := m.Run()
	os.Exit(result)
}

func TestCoreDocumentProcessor_SendNilDocument(t *testing.T) {
	err := cdp.Send(nil, nil, [identity.CentIDByteLength]byte{})
	assert.Error(t, err, "should have thrown an error")
}

func TestCoreDocumentProcessor_AnchorNilDocument(t *testing.T) {
	err := cdp.Anchor(nil, nil)
	assert.Error(t, err, "should have thrown an error")
}
