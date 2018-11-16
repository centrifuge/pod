// +build unit

package documents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	err := Bootstrapper{}.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRegistry])
	_, ok := ctx[BootstrappedRegistry].(*ServiceRegistry)
	assert.True(t, ok)
}
