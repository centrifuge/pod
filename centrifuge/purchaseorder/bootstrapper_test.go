// +build unit

package purchaseorder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}
