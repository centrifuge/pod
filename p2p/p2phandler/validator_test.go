// +build unit

package p2phandler

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/stretchr/testify/assert"
)

func TestValidate_versionValidator(t *testing.T) {
	vv := versionValidator()

	// Nil header
	err := vv.Validate(nil)
	assert.NotNil(t, err)

	// Empty header
	header := &p2ppb.CentrifugeHeader{}
	err = vv.Validate(header)
	assert.NotNil(t, err)

	// Incompatible Major
	header.CentNodeVersion = "1.1.1"
	err = vv.Validate(header)
	assert.NotNil(t, err)

	// Compatible Minor
	header.CentNodeVersion = "0.1.1"
	err = vv.Validate(header)
	assert.Nil(t, err)

	//Same version
	header.CentNodeVersion = version.GetVersion().String()
	err = vv.Validate(header)
	assert.Nil(t, err)
}

func TestValidate_networkValidator(t *testing.T) {
	nv := networkValidator()

	// Nil header
	err := nv.Validate(nil)
	assert.NotNil(t, err)

	// Empty header
	header := &p2ppb.CentrifugeHeader{}
	err = nv.Validate(header)
	assert.NotNil(t, err)

	// Incompatible network
	header.NetworkIdentifier = 12
	err = nv.Validate(header)
	assert.NotNil(t, err)

	// Compatible network
	header.NetworkIdentifier = config.Config.GetNetworkID()
	err = nv.Validate(header)
	assert.Nil(t, err)
}

func TestValidate_handshakeValidator(t *testing.T) {
	hv := handshakeValidator()

	// Incompatible version and network
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   "version",
		NetworkIdentifier: 52,
	}
	err := hv.Validate(header)
	assert.NotNil(t, err)

	// Incompatible version, correct network
	header.NetworkIdentifier = config.Config.GetNetworkID()
	err = hv.Validate(header)
	assert.NotNil(t, err)

	// Compatible version, incorrect network
	header.NetworkIdentifier = 52
	header.CentNodeVersion = version.GetVersion().String()
	err = hv.Validate(header)
	assert.NotNil(t, err)

	// Compatible version and network
	header.NetworkIdentifier = config.Config.GetNetworkID()
	err = hv.Validate(header)
	assert.Nil(t, err)
}
