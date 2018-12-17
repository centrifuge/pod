// +build unit

package healthcheck

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestHandler_Ping(t *testing.T) {
	h := GRPCHandler(cfg)
	pong, err := h.Ping(context.Background(), &empty.Empty{})
	assert.Nil(t, err)
	assert.NotNil(t, pong)
	assert.Equal(t, pong.Version, version.GetVersion().String())
	assert.Equal(t, pong.Network, cfg.GetNetworkString())
}
