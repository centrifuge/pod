// +build ethereum

package ethereum_test

import (
	"os"
	"testing"

	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/magiconair/properties/assert"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestGetConnection_returnsSameConnection(t *testing.T) {
	//TODO this will currently fail if concurrency is at play - e.g. running with 3 go-routines the test will fail
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
