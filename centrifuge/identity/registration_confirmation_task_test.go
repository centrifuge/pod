package identity

import (
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/gocelery"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := RegistrationConfirmationTask{}
	id := tools.RandomSlice32()
	var b32Id [32]byte
	copy(b32Id[:], id[:32])
	decoded, err := simulateJsonDecode(b32Id)
	err = rct.ParseKwargs(decoded)
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

func simulateJsonDecode(b32Id [32]byte) (map[string]interface{}, error) {
	kwargs := map[string]interface{}{CentIdParam: b32Id}
	t1 := gocelery.TaskMessage{Kwargs: kwargs}
	encoded, err := t1.Encode()
	if err != nil {
		return nil, err
	}
	t2, err := gocelery.DecodeTaskMessage(encoded)
	return t2.Kwargs, err
}
