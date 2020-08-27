// +build unit

package configstore

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_BootstrapHappy(t *testing.T) {
	b := Bootstrapper{}
	err := b.Bootstrap(ctx)
	assert.NoError(t, err)
	configService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	assert.True(t, ok)
	_, err = configService.GetConfig()
	assert.NoError(t, err)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	nc := NewNodeConfig(cfg)
	i, err := nc.GetIdentityID()
	assert.NoError(t, err)
	_, err = configService.GetAccount(i)
	assert.NoError(t, err)
}
