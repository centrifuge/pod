// +build unit

package healthcheck

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.InitTestConfig()
	result := m.Run()
	os.Exit(result)
}

func TestService_Ping(t *testing.T) {
	srv := &Handler{}
	pong, err := srv.Ping(context.Background(), new(empty.Empty))
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, pong, "must not be nil")
	assert.Equal(t, pong.Version, version.GetVersion().String())
	assert.Equal(t, pong.Network, config.Config.GetNetworkString())
}
