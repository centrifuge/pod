// +build unit

package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainBootstrapper_BootstrapNoDefaultBootstrappers(t *testing.T) {
	m := &MainBootstrapper{}
	err := m.Bootstrap(map[string]interface{}{})
	assert.Nil(t, err)
}
