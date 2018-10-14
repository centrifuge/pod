// +build unit

package nft_test

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/nft"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&nft.Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}
