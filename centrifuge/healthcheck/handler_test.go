// +build unit

package healthcheck

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestHandler_Ping(t *testing.T) {
	h := GRPCHandler()
	pong, err := h.Ping(context.Background(), &empty.Empty{})
	assert.Nil(t, err)
	assert.NotNil(t, pong)
	assert.Equal(t, pong.Version, version.GetVersion().String())
	assert.Equal(t, pong.Network, config.Config.GetNetworkString())
}
