// +build unit

package context

import (
	"testing"

	testing2 "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/stretchr/testify/assert"
)

func TestMainBootstrapper_Bootstrap(t *testing.T) {
	testing2.InitTestConfig()
	testing2.InitTestStoragePath()
	m := &MainBootstrapper{}
	m.PopulateDefaultBootstrappers()
	err := m.Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an Ethereum connection error")
}

func TestMainBootstrapper_BootstrapNoDefaultBootstrappers(t *testing.T) {
	m := &MainBootstrapper{}
	err := m.Bootstrap(map[string]interface{}{})
	assert.Nil(t, err)
}
