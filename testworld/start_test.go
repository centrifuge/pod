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
	multiHostMultiAccount          = "multiHostMultiAccount"
)

var isRunningOnCI = len(os.Getenv("TRAVIS")) != 0

// Adjust these based on local testing requirments, please revert for CI server
var configFile = "configs/local.json"
var runPOAGeth = !isRunningOnCI

// make this true this when running for the first time in local env
var createHostConfigs = isRunningOnCI

// make this false if you want to make the tests run faster locally, but revert before committing to repo
var runMigrations = !isRunningOnCI

// doctorFord manages the hosts
var doctorFord *hostManager

func TestMain(m *testing.M) {
	c, err := loadConfig(configFile)
	if err != nil {
		panic(err)
	}
	if runPOAGeth {
		// NOTE that we don't bring down geth automatically right now because this must only be used for local testing purposes
		startPOAGeth()
	}
	if runMigrations {
		runSmartContractMigrations()
	}
	var contractAddresses *config.SmartContractAddresses
	if c.Network == "testing" {
		contractAddresses = getSmartContractAddresses()
	}
	doctorFord = newHostManager(c.EthNodeURL, c.AccountKeyPath, c.AccountPassword, c.Network, c.TxPoolAccess, contractAddresses)
	err = doctorFord.init(createHostConfigs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
}
