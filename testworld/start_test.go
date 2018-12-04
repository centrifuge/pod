// This is the starting point for all Testworld tests

package testworld

import (
	"fmt"
	"os"
	"testing"
)

var isRunningOnCI = len(os.Getenv("TRAVIS")) != 0

// Adjust these based on local testing requirments, please revert for CI server
var configFile = "configs/local.json"
var runPOAGeth = !isRunningOnCI

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
	contractAddresses := getSmartContractAddresses()
	doctorFord = newHostManager(c.EthNodeUrl, c.AccountKeyPath, c.AccountPassword, c.Network, c.TxPoolAccess, contractAddresses)
	err = doctorFord.init()
	if err != nil {
		panic(err)
	}
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
}
