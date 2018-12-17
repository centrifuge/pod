// +build unit

package nft

import (
	"encoding/hex"
	"testing"

	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestMintingConfirmationTask_ParseKwargs_success(t *testing.T) {
	task := mintingConfirmationTask{}
	tokenId := hex.EncodeToString(utils.RandomSlice(256))
	blockHeight := uint64(12)
	registryAddress := "0xf72855759a39fb75fc7341139f5d7a3974d4da08"

	kwargs := map[string]interface{}{
		tokenIDParam:           tokenId,
		queue.BlockHeightParam: blockHeight,
		registryAddressParam:   registryAddress,
	}

	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Nil(t, err, "parsing should be successful")

	assert.Equal(t, tokenId, task.TokenID, "tokenId should be parsed correctly")
	assert.Equal(t, blockHeight, task.BlockHeight, "blockHeight should be parsed correctly")
	assert.Equal(t, registryAddress, task.RegistryAddress, "registryAddress should be parsed correctly")

}

func TestMintingConfirmationTask_ParseKwargs_fail(t *testing.T) {
	task := mintingConfirmationTask{}
	tests := []map[string]interface{}{
		{
			queue.BlockHeightParam: uint64(12),
			registryAddressParam:   "0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		},
		{
			tokenIDParam:         hex.EncodeToString(utils.RandomSlice(256)),
			registryAddressParam: "0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		},
		{
			tokenIDParam:           hex.EncodeToString(utils.RandomSlice(256)),
			queue.BlockHeightParam: uint64(12),
		},
		{
			//empty map

		},
		{
			"dummy": "dummy",
		},
	}

	for i, test := range tests {
		decoded, err := utils.SimulateJSONDecodeForGocelery(test)
		assert.Nil(t, err, "json decode should not thrown an error")
		err = task.ParseKwargs(decoded)
		assert.Error(t, err, "test case %v: parsing should fail", i)
	}
}
