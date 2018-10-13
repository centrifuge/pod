// +build unit

package nft_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/centrifuge/go-centrifuge/centrifuge/nft"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&nft.Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}
