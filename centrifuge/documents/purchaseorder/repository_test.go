// +build unit

package purchaseorder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_getRepository(t *testing.T) {
	r := getRepository()
	assert.NotNil(t, r)
	assert.Equal(t, "purchaseorder", r.(*repository).KeyPrefix)
}
