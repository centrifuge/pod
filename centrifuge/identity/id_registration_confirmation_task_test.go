// +build unit

package identity_test

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := identity.IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(identity.CentIDLength)
	blockHeight := uint64(3132)
	idBytes, _ := identity.ToCentID(id)
	kwargs := map[string]interface{}{
		identity.CentIdParam: idBytes,
		identity.BlockHeight: blockHeight,
	}
	decoded, err := tools.SimulateJsonDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("Could not parse %s for [%x]", identity.CentIdParam, id)
	}
	assert.Equal(t, idBytes, rct.CentID, "Resulting mockID should have the same ID as the input")
	assert.Equal(t, blockHeight, rct.BlockHeight, "Resulting blockheight should be same as the input")
}

func TestRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := identity.IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(identity.CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{"notId": id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestRegistrationConfirmationTask_ParseKwargsInvalidType(t *testing.T) {
	rct := identity.IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(identity.CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{identity.CentIdParam: id})
	assert.NotNil(t, err, "Should not parse without the correct type of centId")
}
