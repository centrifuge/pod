// +build unit

package pedersen

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestHashPayload(t *testing.T) {
	t.SkipNow() //Skipping until travis is configured properly
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
