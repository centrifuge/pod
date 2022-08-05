// This is the starting point for all Testworld tests
//go:build testworld

package testworld

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"

	"github.com/centrifuge/go-centrifuge/centchain"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/cmd"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/config"
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

var testworldBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{}, // required for starting Centrifuge chain
	&config.Bootstrapper{},           // required for identity V2 bootstrapper
	&leveldb.Bootstrapper{},          // required for identity V2 bootstrapper
	jobs.Bootstrapper{},              // required for identity V2 bootstrapper
	&configstore.Bootstrapper{},      // required for identity V2 bootstrapper
	&dispatcher.Bootstrapper{},       // required for identity V2 bootstrapper
	centchain.Bootstrapper{},         // required for identity V2 bootstrapper
	&v2.Bootstrapper{},               // required for creating the proxies for Alice
}

func cleanUpTestBootstrapContext(ctx map[string]any) error {
	db, ok := ctx[storage.BootstrappedDB].(storage.Repository)

	if !ok {
		return errors.New("bootstrapped DB not initialised")
	}

	_ = db.Close()

	db, ok = ctx[storage.BootstrappedConfigDB].(storage.Repository)

	if !ok {
		return errors.New("bootstrapped config DB not initialised")
	}

	_ = db.Close()

	return nil
}

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelDebug)
	//err := setMaxLimits()
	//if err != nil {
	//	log.Warn(err)
	//}
	//
	//network := os.Getenv(networkEnvKey)
	//if network == "" {
	//	network = "testing"
	//}

	// Create config file.
	cfgVals, err := getConfigVals()

	if err != nil {
		panic(fmt.Errorf("couldn't retrieve config vals: %w", err))
	}

	cfgFile, err := config.CreateConfigFile(cfgVals)

	if err != nil {
		panic(fmt.Errorf("couldn't create config file: %w", err))
	}

	cfg := config.LoadConfiguration(cfgFile.ConfigFileUsed())

	// Generate the keys required to run the node.

	err = cmd.GenerateNodeKeys(cfg)

	if err != nil {
		panic(fmt.Errorf("couldn't generate node keys: %w", err))
	}

	// Create test bootstrap context with config file.

	testBootstrapCtx := map[string]any{
		config.BootstrappedConfigFile: cfgFile.ConfigFileUsed(),
	}

	// Run bootstrappers that start the Centrifuge chain and create the accounts used for testing.

	testBootstrapCtx = bootstrap.RunTestBootstrappers(testworldBootstrappers, testBootstrapCtx)

	if err := cleanUpTestBootstrapContext(testBootstrapCtx); err != nil {
		panic(fmt.Errorf("couldn't cleanup test bootstrap context: %w", err))
	}

	// Create bootstrap context with config file.

	serviceCtx := map[string]any{
		config.BootstrappedConfigFile: cfgFile.ConfigFileUsed(),
	}

	// Run the base bootstrappers

	mainBootstrapper := bootstrappers.MainBootstrapper{}
	mainBootstrapper.PopulateBaseBootstrappers()

	err = mainBootstrapper.Bootstrap(serviceCtx)

	if err != nil {
		panic(fmt.Errorf("couldn't bootstrap the node"))
	}

	//c, err := loadConfig(network)
	//if err != nil {
	//	panic(err)
	//}

	// run geth and cent chain
	//if c.RunNetwork {
	//	testingutils.StartCentChain()
	//}

	// run migrations if required
	//if c.RunMigrations {
	//	testingutils.RunSmartContractMigrations()
	//}

	// run bridge
	//if c.RunNetwork {
	//	testingutils.StartBridge()
	//}

	//if c.Network == "testing" {
	//	c.ContractAddresses = testingutils.GetSmartContractAddresses()
	//	c.DappAddresses = testingutils.GetDAppSmartContractAddresses()
	//}
	//
	//err = setEthEnvKeys(c.EthAccountKeyPath, c.EthAccountPassword)
	//if err != nil {
	//	panic(err)
	//}

	doctorFord = newHostManager(cfg, cfgFile.ConfigFileUsed(), serviceCtx)

	err = doctorFord.init()
	if err != nil {
		panic(err)
	}
	//fmt.Printf("contract addresses %+v\n", c.ContractAddresses)
	//fmt.Printf("Dapp contract addresses %+v\n", c.DappAddresses)
	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
}

func setMaxLimits() error {
	if isRunningOnCI {
		//log.Debug("Running on CI. Not setting limits")
		return nil
	}

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	//log.Debugf("Previous Rlimits: %v", rLimit)
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

	//log.Debugf("Current Rlimits: %v", rLimit)
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
