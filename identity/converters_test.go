// +build unit

package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConverters(t *testing.T) {
	sdids := []string{
		"0xF72855759A39FB75fC7341139f5d7A3974d4DA08",
		"0xF72855759A39Fb75fc7341139F5d7a3974C4dA08",
		"",
	}

	dids, err := StringsToDIDs(sdids...)
	assert.NoError(t, err)

	bdids := DIDsToBytes(dids...)
	cdids := BytesToDIDs(bdids...)

	esdids := DIDsToStrings(cdids...)
	assert.Equal(t, sdids, esdids)

	sdids[2] = "wrong id"
	_, err = StringsToDIDs(sdids...)
	assert.Error(t, err)
}
