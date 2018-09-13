// +build unit

package coredocumentprocessor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
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
	err := cdp.Anchor(nil, nil, []identity.CentID{})
	assert.Error(t, err, "should have thrown an error")
}
