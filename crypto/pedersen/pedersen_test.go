// +build pedersenunit

package pedersen

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestHashPayload(t *testing.T) {
	ph := New()

	_, err := ph.Write(utils.RandomSlice(64))
	assert.NoError(t, err)

	s := ph.Sum(nil)
	assert.Len(t, s, 32)

	ph.Reset()
	// wrong length
	_, err = ph.Write(utils.RandomSlice(32))
	assert.Error(t, err)
}
