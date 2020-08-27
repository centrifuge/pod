// +build unit

package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateValidator(t *testing.T) {
	uv := UpdateValidator(nil)
	assert.Len(t, uv, 1)
}
