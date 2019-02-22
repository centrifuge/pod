package transactions

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestTxID_String(t *testing.T) {
	tID := NewTxID()
	tidStr := tID.String()
	bs, err := hexutil.Decode(tidStr)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(tID[:], bs))
	tIDConv, err := FromString(tidStr)
	assert.NoError(t, err)
	assert.True(t, TxIDEqual(tID, tIDConv))
}
