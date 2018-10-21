// +build integration

package ethereum_test

import (
	"os"
	"testing"

	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.DONT_USE_FOR_UNIT_TESTS_TestFunctionalEthereumBootstrap()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestGetConnection_returnsSameConnection(t *testing.T) {
	howMany := 5
	confChannel := make(chan ethereum.EthereumClient, howMany)
	for ix := 0; ix < howMany; ix++ {
		go func() {
			confChannel <- ethereum.GetConnection()
		}()
	}
	for ix := 0; ix < howMany; ix++ {
		multiThreadCreatedCon := <-confChannel
		assert.Equal(t, multiThreadCreatedCon, ethereum.GetConnection(), "Should only return a single ethereum client")
	}
}
