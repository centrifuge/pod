package jobs

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestJobID_String(t *testing.T) {
	id := NewJobID()
	idStr := id.String()
	bs, err := hexutil.Decode(idStr)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(id[:], bs))
	idConv, err := FromString(idStr)
	assert.NoError(t, err)
	assert.True(t, JobIDEqual(id, idConv))
}

func TestNilJobID(t *testing.T) {
	id := NewJobID()
	assert.NotEmpty(t, id.String())

	id = NilJobID()
	assert.Empty(t, id.String())
}
