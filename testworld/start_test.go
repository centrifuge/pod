// This is the starting point for all Testworld tests
// +build testworld

package testworld

import (
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
)

type testType string

const (
	withinHost            testType = "withinHost"
	multiHost             testType = "multiHost"
	multiHostMultiAccount testType = "multiHostMultiAccount"
)

var isRunningOnCI = len(os.Getenv("TRAVIS")) != 0

// doctorFord manages the hosts
var doctorFord *hostManager

func TestMain(m *testing.M) {
	c, configName, err := loadConfig(!isRunningOnCI)
	if err != nil {
		panic(err)
	}
	if c.RunPOAGeth {
		// NOTE that we don't bring down geth automatically right now because this must only be used for local testing purposes
		startPOAGeth()
	}
	if c.RunMigrations {
		runSmartContractMigrations()
	}
	var contractAddresses *config.SmartContractAddresses
	if c.Network == "testing" {
		contractAddresses = getSmartContractAddresses()
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
