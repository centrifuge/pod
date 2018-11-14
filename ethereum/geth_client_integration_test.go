// +build integration

package ethereum_test

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/stretchr/testify/assert"
)

var cfg *config.Configuration

func TestMain(m *testing.M) {
	ctx := cc.TestFunctionalEthereumBootstrap()
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestGetConnection_returnsSameConnection(t *testing.T) {
	howMany := 5
	confChannel := make(chan ethereum.Client, howMany)
	for ix := 0; ix < howMany; ix++ {
		go func() {
			confChannel <- ethereum.GetClient()
		}()
	}
	for ix := 0; ix < howMany; ix++ {
		multiThreadCreatedCon := <-confChannel
		assert.Equal(t, multiThreadCreatedCon, ethereum.GetClient(), "Should only return a single ethereum client")
	}
}

func TestNewGethClient(t *testing.T) {
	gc, err := ethereum.NewGethClient(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, gc)
}
