// This is the starting point for all Testworld tests
// +build testworld

package testworld

import (
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/testingutils"
)

type testType string

const (
	withinHost            testType = "withinHost"
	multiHost             testType = "multiHost"
	multiHostMultiAccount testType = "multiHostMultiAccount"
)

// doctorFord manages the hosts
var doctorFord *hostManager

func TestMain(m *testing.M) {
	c, configName, err := loadConfig(!isRunningOnCI)
	if err != nil {
		panic(err)
	}
	if c.RunChains {
		// NOTE that we don't bring down geth/cc automatically right now because this must only be used for local testing purposes
		testingutils.StartPOAGeth()
		testingutils.StartCentChain()
	}
	if c.RunMigrations {
		testingutils.RunSmartContractMigrations()
	}
	var contractAddresses *config.SmartContractAddresses
	if c.Network == "testing" {
		contractAddresses = testingutils.GetSmartContractAddresses()
	}
	doctorFord = newHostManager(c.EthNodeURL, c.AccountKeyPath, c.AccountPassword, c.Network, configName, c.TxPoolAccess, contractAddresses)
	err = doctorFord.init(c.CreateHostConfigs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
}
