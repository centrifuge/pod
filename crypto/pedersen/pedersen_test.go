// +build unit

package pedersen

import (
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashPayload(t *testing.T) {
	ph := NewPedersenHash()

	_, err := ph.Write(utils.RandomSlice(64))
	assert.NoError(t, err)

	s := ph.Sum(nil)
	assert.Len(t, s, 32)

	ph.Reset()
	// wrong length
	_, err = ph.Write(utils.RandomSlice(32))
	assert.Error(t, err)
}