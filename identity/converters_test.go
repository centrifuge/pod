//go:build unit

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
	cdids, err := BytesToDIDs(bdids...)
	assert.NoError(t, err)

	esdids := DIDsToStrings(cdids...)
	assert.Equal(t, sdids, esdids)

	sdids[2] = "wrong id"
	_, err = StringsToDIDs(sdids...)
	assert.Error(t, err)
}

func TestDIDsPointers(t *testing.T) {
	sdids := []string{
		"0xF72855759A39FB75fC7341139f5d7A3974d4DA08",
		"0xF72855759A39Fb75fc7341139F5d7a3974C4dA08",
		"",
	}

	pdids, err := StringsToDIDs(sdids...)
	assert.NoError(t, err)
	assert.Len(t, pdids, 3)

	vdids := FromPointerDIDs(pdids...)
	assert.Len(t, vdids, 3)

	epdids := DIDsPointers(vdids...)
	assert.Equal(t, pdids, epdids)
}

func TestRemoveDuplicateDIDs(t *testing.T) {
	sdids := []string{
		"0xF72855759A39FB75fC7341139f5d7A3974d4DA08",
		"0xF72855759A39Fb75fc7341139F5d7a3974C4dA08",
		"0xF72855759A39Fb75fc7341139F5d7a3974C4dA08",
	}

	pdids, err := StringsToDIDs(sdids...)
	assert.NoError(t, err)
	assert.Len(t, pdids, 3)

	dids := FromPointerDIDs(pdids...)
	assert.Len(t, dids, 3)

	ddids := RemoveDuplicateDIDs(dids)
	assert.Len(t, ddids, 2)
}
