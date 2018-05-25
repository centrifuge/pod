// +build ethereum

package ethereum_test

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/magiconair/properties/assert"
	"os"
	"testing"
)

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
