// +build unit

package api

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
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
	m[configstore.BootstrappedConfigStorage] = new(configstore.MockService)
	m[documents.BootstrappedRegistry] = documents.NewServiceRegistry()
	err = b.Bootstrap(m)
	assert.Nil(t, err)
	assert.NotNil(t, m[bootstrap.BootstrappedAPIServer])
	_, ok := m[bootstrap.BootstrappedAPIServer].(node.Server)
	assert.True(t, ok)
}
