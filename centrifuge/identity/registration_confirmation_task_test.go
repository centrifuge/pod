package identity

import (
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := RegistrationConfirmationTask{}
	id := tools.RandomSlice32()
	var b32Id [32]byte
	copy(b32Id[:], id[:32])
	err := rct.ParseKwargs(map[string]interface{}{CentIdParam: b32Id})
	if err != nil {
		t.Errorf("Could not parse %s for [%x]", CentIdParam, id)
	}
	assert.Equal(t, b32Id, rct.CentId, "Resulting id should have the same ID as the input")
}

func TestRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := RegistrationConfirmationTask{}
	id := tools.RandomSlice32()
	var b32Id [32]byte
	copy(b32Id[:], id[:32])
	err := rct.ParseKwargs(map[string]interface{}{"notId": b32Id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestRegistrationConfirmationTask_ParseKwargsMalformedCentId(t *testing.T) {
	rct := RegistrationConfirmationTask{}
	id := tools.RandomSlice32()
	err := rct.ParseKwargs(map[string]interface{}{CentIdParam: id})
	assert.NotNil(t, err, "Should not parse CentId of type other than [32]byte")
}
