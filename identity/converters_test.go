// +build unit

package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConverters(t *testing.T) {
	sdids := []string{
		"0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		"0xf72855759a39fb75fc7341139f5d7a3974c4da08",
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
