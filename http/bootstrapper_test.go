//go:build unit

package http

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/node"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	b := Bootstrapper{}

	// no config
	m := make(map[string]interface{})
	err := b.Bootstrap(m)
	assert.Error(t, err)

	// config
	m[bootstrap.BootstrappedConfig] = new(testingconfig.MockConfig)
	cs := new(configstore.MockService)
	m[config.BootstrappedConfigStorage] = cs
	cs.On("GetConfig").Return(&configstore.NodeConfig{}, nil)
	err = b.Bootstrap(m)
	assert.Nil(t, err)
	assert.NotNil(t, m[bootstrap.BootstrappedAPIServer])
	_, ok := m[bootstrap.BootstrappedAPIServer].(node.Server)
	assert.True(t, ok)
}
