// +build unit

package identity

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := idRegistrationConfirmationTask{}
	id := utils.RandomSlice(CentIDLength)
	blockHeight := uint64(3132)
	timeout := float64(3000)
	idBytes, _ := ToCentID(id)
	kwargs := map[string]interface{}{
		centIDParam:            idBytes,
		queue.BlockHeightParam: blockHeight,
		queue.TimeoutParam:     timeout,
	}
	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("Could not parse %s for [%x]", centIDParam, id)
	}
	assert.Equal(t, idBytes, rct.centID, "Resulting mockID should have the same ID as the input")
	assert.Equal(t, blockHeight, rct.blockHeight, "Resulting blockheight should be same as the input")
}

func TestRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := idRegistrationConfirmationTask{}
	id := utils.RandomSlice(CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{"notId": id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestRegistrationConfirmationTask_ParseKwargsInvalidType(t *testing.T) {
	rct := idRegistrationConfirmationTask{}
	id := utils.RandomSlice(CentIDLength)
	err := rct.ParseKwargs(map[string]interface{}{centIDParam: id})
	assert.NotNil(t, err, "Should not parse without the correct type of centId")
}
