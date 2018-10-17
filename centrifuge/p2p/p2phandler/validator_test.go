// +build unit

package p2phandler

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
)

func TestValidate_versionValidator(t *testing.T) {
	vv := versionValidator()

	// Empty version
	err := vv.Validate("")
	assert.NotNil(t, err)

	// Wrong version
	err = vv.Validate(34)
	assert.NotNil(t, err)

	// Incompatible Major
	err = vv.Validate("1.1.1")
	assert.NotNil(t, err)

	// Compatible Minor
	err = vv.Validate("0.1.1")
	assert.Nil(t, err)

	//Same version
	err = vv.Validate(version.GetVersion().String())
	assert.Nil(t, err)
}

func TestValidate_networkValidator(t *testing.T) {
	nv := networkValidator()

	// Empty network
	err := nv.Validate(nil)
	assert.NotNil(t, err)

	// Wrong network
	err = nv.Validate("blabla")
	assert.NotNil(t, err)

	// Incompatible network
	err = nv.Validate(12)
	assert.NotNil(t, err)

	// Compatible network
	err = nv.Validate(config.Config.GetNetworkID())
	assert.Nil(t, err)
}

func TestValidate_handshakeValidator(t *testing.T) {
	hv := handshakeValidator()

	// Mismatch number of parameters
	err := hv.Validate([]interface{}{"version"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Mismatched validator params, expected [2], actual [1]")

	// Incompatible version and network
	err = hv.Validate([]interface{}{"version", 52})
	assert.NotNil(t, err)

	// Incompatible version, correct network
	err = hv.Validate([]interface{}{"version", config.Config.GetNetworkID()})
	assert.NotNil(t, err)

	// Compatible version, incorrect network
	err = hv.Validate([]interface{}{version.GetVersion().String(), 52})
	assert.NotNil(t, err)

	// Compatible version and network
	err = hv.Validate([]interface{}{version.GetVersion().String(), config.Config.GetNetworkID()})
	assert.Nil(t, err)
}
