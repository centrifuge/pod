package transactions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxID_String(t *testing.T) {
	tID := NewTxID()
	assert.Equal(t, "", tID.String())
}
