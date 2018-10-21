// +build unit

package identity_test

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/stretchr/testify/assert"
)

func TestKeyRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := identity.KeyRegistrationConfirmationTask{}
	id := utils.RandomSlice(identity.CentIDLength)
	key := utils.RandomSlice(32)
	var keyFixed [32]byte
	copy(keyFixed[:], key)
	keyPurpose := identity.KeyPurposeSigning
	blockHeight := uint64(12)
	idBytes, _ := identity.ToCentID(id)
	kwargs := map[string]interface{}{
		identity.CentIdParam:     idBytes,
		identity.KeyParam:        keyFixed,
		identity.KeyPurposeParam: keyPurpose,
		identity.BlockHeight:     blockHeight,
	}
	decoded, err := utils.SimulateJsonDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("parse error %s", err.Error())
	}
	assert.Equal(t, idBytes, rct.CentID, "Resulting mockID should have the same ID as the input")
	assert.Equal(t, keyFixed, rct.Key, "Resulting key should be same as the input")
	assert.Equal(t, keyPurpose, rct.KeyPurpose, "Resulting keyPurpose should be same as the input")
	assert.Equal(t, blockHeight, rct.BlockHeight, "Resulting blockheight should be same as the input")
}

func TestKeyRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := identity.KeyRegistrationConfirmationTask{}
	id := utils.RandomSlice(identity.CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{"notId": id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestKeyRegistrationConfirmationTask_ParseKwargsInvalidType(t *testing.T) {
	rct := identity.KeyRegistrationConfirmationTask{}
	id := utils.RandomSlice(identity.CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{identity.CentIdParam: id})
	assert.NotNil(t, err, "Should not parse without the correct type of centId")
}
