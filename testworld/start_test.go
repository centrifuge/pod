// This is the starting point for all Testworld tests
//go:build testworld
// +build testworld

package testworld

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils"
	logging "github.com/ipfs/go-log"
)

type testType string

const (
	withinHost            testType = "withinHost"
	multiHost             testType = "multiHost"
	multiHostMultiAccount testType = "multiHostMultiAccount"

	networkEnvKey = "TESTWORLD_NETWORK"
)

// doctorFord manages the hosts
var doctorFord *hostManager

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelInfo)
	err := setMaxLimits()
	if err != nil {
		log.Warn(err)
	}

	network := os.Getenv(networkEnvKey)
	if network == "" {
		network = "testing"
	}

	c, err := loadConfig(network)
	if err != nil {
		panic(err)
	}

	// run geth and cent chain
	if c.RunNetwork {
		testingutils.StartPOAGeth()
		testingutils.StartCentChain()
	}

	// run migrations if required
	if c.RunMigrations {
		testingutils.RunSmartContractMigrations()
	}

	// run bridge
	if c.RunNetwork {
		testingutils.StartBridge()
	}

	if c.Network == "testing" {
		c.ContractAddresses = testingutils.GetSmartContractAddresses()
		c.DappAddresses = testingutils.GetDAppSmartContractAddresses()
	}

	err = setEthEnvKeys(c.EthAccountKeyPath, c.EthAccountPassword)
	if err != nil {
		panic(err)
	}

	doctorFord = newHostManager(c)
	err = doctorFord.init()
	if err != nil {
		panic(err)
	}
	fmt.Printf("contract addresses %+v\n", c.ContractAddresses)
	fmt.Printf("Dapp contract addresses %+v\n", c.DappAddresses)
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
		return fmt.Errorf("error setting Rlimit: %w", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("error getting Rlimit: %w", err)
	}

	log.Debugf("Current Rlimits: %v", rLimit)
	return nil
}

func setEthEnvKeys(path, password string) error {
	bfile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = os.Setenv("CENT_ETHEREUM_ACCOUNTS_MAIN_KEY", string(bfile))
	if err != nil {
		return err
	}

	return os.Setenv("CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD", password)
}
