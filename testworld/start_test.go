// This is the starting point for all Testworld tests
// +build testworld

package testworld

import (
	"fmt"
	"os"
	"syscall"
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
	err := setMaxLimits()
	if err != nil {
		panic(err)
	}
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
	dappAddresses := make(map[string]string)
	if c.Network == "testing" {
		contractAddresses = testingutils.GetSmartContractAddresses()
		testingutils.RunDAppSmartContractMigrations()
		dappAddresses = testingutils.GetDAppSmartContractAddresses()
	}
	doctorFord = newHostManager(
		c.EthNodeURL, c.AccountKeyPath, c.AccountPassword, c.Network, configName, c.TxPoolAccess, contractAddresses, dappAddresses)
	err = doctorFord.init(c.CreateHostConfigs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
}

func setMaxLimits() error {
	if isRunningOnCI {
		log.Debug("Running on CI. Not setting limits")
		return nil
	}

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	log.Debugf("Previous Rlimits: %v", rLimit)
	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("error setting Rlimit: %v", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("error getting Rlimit: %v", err)
	}

	log.Debugf("Current Rlimits: %v", rLimit)
	return nil
}
