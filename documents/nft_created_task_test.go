// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestNftCreatedTask_ParseKwargs(t *testing.T) {
	task := new(nftCreatedTask)
	m := make(map[string]interface{})

	// no transactions
	assert.Error(t, task.ParseKwargs(m))

	// no account
	m[transactions.TxIDParam] = uuid.Must(uuid.NewV4()).String()
	assert.Error(t, task.ParseKwargs(m))

	// wrong format
	m[AccountIDParam] = "0x1002030405"
	assert.Error(t, task.ParseKwargs(m))

	// no document
	m[AccountIDParam] = "0x010203040506"
	assert.Error(t, task.ParseKwargs(m))

	// hex fails
	m[DocumentIDParam] = "sfkvfj"
	assert.Error(t, task.ParseKwargs(m))

	// missing registry
	m[DocumentIDParam] = hexutil.Encode(utils.RandomSlice(32))
	assert.Error(t, task.ParseKwargs(m))

	// missing token ID
	m[TokenRegistryParam] = "0xf72855759a39fb75fc7341139f5d7a3974d4da08"
	assert.Error(t, task.ParseKwargs(m))

	// hex util fails
	m[TokenIDParam] = "lkfv"
	assert.Error(t, task.ParseKwargs(m))

	m[TokenIDParam] = hexutil.Encode(utils.RandomSlice(32))
	assert.NoError(t, task.ParseKwargs(m))
}
