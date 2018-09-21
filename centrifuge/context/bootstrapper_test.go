// +build unit

package context

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	testing2 "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/stretchr/testify/assert"
)

func TestMainBootstrapper_Bootstrap(t *testing.T) {
	testing2.InitTestConfig()
	testing2.InitTestStoragePath()
	// set a dummy url here so that ethereum will always fail to connect
	config.Config.V.Set("ethereum.nodeURL", "blah")
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
