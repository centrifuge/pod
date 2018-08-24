package identity

import (
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationConfirmationTask_ParseKwargs(t *testing.T) {
	rct := RegistrationConfirmationTask{}
	id := tools.RandomSlice32()
	err := rct.ParseKwargs(map[string]interface{}{CentIdParam: id})
	if err != nil {
		t.Errorf("Could not parse %s for [%x]", CentIdParam, id)
	}
	assert.Equal(t, id, rct.CentId[:], "Resulting id should have the same ID as the input")
}
