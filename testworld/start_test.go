//go:build testworld

package testworld

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/storage"
	logging "github.com/ipfs/go-log"
)

// doctorFord manages the hosts
var doctorFord *hostManager

var testworldBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{}, // required for starting Centrifuge chain
	//&config.Bootstrapper{},           // required for identity V2 bootstrapper
	//&leveldb.Bootstrapper{},          // required for identity V2 bootstrapper
	//jobs.Bootstrapper{},              // required for identity V2 bootstrapper
	//&configstore.Bootstrapper{},      // required for identity V2 bootstrapper
	//&dispatcher.Bootstrapper{},       // required for identity V2 bootstrapper
	//centchain.Bootstrapper{},         // required for identity V2 bootstrapper
	//&v2.Bootstrapper{},               // required for creating the proxies for Alice
}

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelDebug)

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

	err = config.GenerateNodeKeys(cfg)

	if err != nil {
		panic(fmt.Errorf("couldn't generate node keys: %w", err))
	}

	// Create test bootstrap context with config file.

	testBootstrapCtx := map[string]any{
		config.BootstrappedConfigFile: cfgFile.ConfigFileUsed(),
	}

	// Run bootstrappers that start the Centrifuge chain and create the accounts used for testing.

	testBootstrapCtx = bootstrap.RunTestBootstrappers(testworldBootstrappers, testBootstrapCtx)

	//if err := cleanUpTestBootstrapContext(testBootstrapCtx); err != nil {
	//	panic(fmt.Errorf("couldn't cleanup test bootstrap context: %w", err))
	//}

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

	doctorFord, err = newHostManager(cfg, cfgFile.ConfigFileUsed(), serviceCtx, testAccountMap)

	if err != nil {
		panic(fmt.Errorf("couldn't create new host manager: %w", err))
	}

	err = doctorFord.init()

	if err != nil {
		panic(fmt.Errorf("couldn't init host manager: %w", err))
	}

	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
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
