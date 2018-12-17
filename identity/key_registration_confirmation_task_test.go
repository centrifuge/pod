// +build unit

package identity

import (
	"testing"

	"time"

	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestKeyRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := keyRegistrationConfirmationTask{timeout: time.Second * 10}
	id := utils.RandomSlice(CentIDLength)
	key := utils.RandomSlice(32)
	var keyFixed [32]byte
	copy(keyFixed[:], key)
	keyPurpose := KeyPurposeSigning
	bh := uint64(12)
	idBytes, _ := ToCentID(id)
	kwargs := map[string]interface{}{
		centIDParam:            idBytes,
		keyParam:               keyFixed,
		keyPurposeParam:        keyPurpose,
		queue.BlockHeightParam: bh,
	}
	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("parse error %s", err.Error())
	}
	assert.Equal(t, idBytes, rct.centID, "Resulting mockID should have the same ID as the input")
	assert.Equal(t, keyFixed, rct.key, "Resulting key should be same as the input")
	assert.Equal(t, keyPurpose, rct.keyPurpose, "Resulting keyPurpose should be same as the input")
	assert.Equal(t, bh, rct.blockHeight, "Resulting blockheight should be same as the input")
}

func TestKeyRegistrationConfirmationTask_ParseKwargsHappyPathOverrideTimeout(t *testing.T) {
	rct := keyRegistrationConfirmationTask{timeout: time.Second * 10}
	id := utils.RandomSlice(CentIDLength)
	key := utils.RandomSlice(32)
	var keyFixed [32]byte
	copy(keyFixed[:], key)
	keyPurpose := KeyPurposeSigning
	bh := uint64(12)
	idBytes, _ := ToCentID(id)
	overrideTimeout := float64(time.Second * 3)
	kwargs := map[string]interface{}{
		centIDParam:            idBytes,
		keyParam:               keyFixed,
		keyPurposeParam:        keyPurpose,
		queue.BlockHeightParam: bh,
		queue.TimeoutParam:     overrideTimeout,
	}
	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("parse error %s", err.Error())
	}
	assert.Equal(t, idBytes, rct.centID, "Resulting mockID should have the same ID as the input")
	assert.Equal(t, keyFixed, rct.key, "Resulting key should be same as the input")
	assert.Equal(t, keyPurpose, rct.keyPurpose, "Resulting keyPurpose should be same as the input")
	assert.Equal(t, bh, rct.blockHeight, "Resulting blockheight should be same as the input")
	assert.Equal(t, time.Duration(overrideTimeout).Seconds(), rct.timeout.Seconds(), "Resulting timeout should be overwritten by kwargs")
}

func TestKeyRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := keyRegistrationConfirmationTask{}
	id := utils.RandomSlice(CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{"notId": id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestKeyRegistrationConfirmationTask_ParseKwargsInvalidType(t *testing.T) {
	rct := keyRegistrationConfirmationTask{}
	id := utils.RandomSlice(CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{centIDParam: id})
	assert.NotNil(t, err, "Should not parse without the correct type of centId")
}
