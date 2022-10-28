//go:build testworld

package testworld

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/testworld/park/behavior"

	"github.com/centrifuge/go-centrifuge/storage"
	logging "github.com/ipfs/go-log"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelDebug)

	head, err := behavior.NewHead()

	if err != nil {
		panic(fmt.Errorf("couldn't create new head: %w", err))
	}

	if err := head.Start(); err != nil {
		panic(fmt.Errorf("couldn't start head: %w", err))
	}

	if err := head.Stop(); err != nil {
		panic(fmt.Errorf("couldn't stop head: %w", err))
	}

	//// Create config file.
	//cfgVals, err := getConfigVals()
	//
	//if err != nil {
	//	panic(fmt.Errorf("couldn't retrieve config vals: %w", err))
	//}
	//
	//cfgFile, err := config.CreateConfigFile(cfgVals)
	//
	//if err != nil {
	//	panic(fmt.Errorf("couldn't create config file: %w", err))
	//}
	//
	//cfg := config.LoadConfiguration(cfgFile.ConfigFileUsed())
	//
	//// Generate the keys required to run the node.
	//
	//err = config.GenerateAndWriteP2PKeys(cfg)
	//
	//if err != nil {
	//	panic(fmt.Errorf("couldn't generate node keys: %w", err))
	//}
	//
	//// Create test bootstrap context with config file.
	//
	//testBootstrapCtx := map[string]any{
	//	config.BootstrappedConfigFile: cfgFile.ConfigFileUsed(),
	//}
	//
	//// Run bootstrappers that start the Centrifuge chain and create the accounts used for testing.
	//
	//testBootstrapCtx = bootstrap.RunTestBootstrappers(testworldBootstrappers, testBootstrapCtx)
	//
	////if err := cleanUpTestBootstrapContext(testBootstrapCtx); err != nil {
	////	panic(fmt.Errorf("couldn't cleanup test bootstrap context: %w", err))
	////}
	//
	//// Create bootstrap context with config file.
	//
	//serviceCtx := map[string]any{
	//	config.BootstrappedConfigFile: cfgFile.ConfigFileUsed(),
	//}
	//
	//// Run the base bootstrappers
	//
	//mainBootstrapper := bootstrappers.MainBootstrapper{}
	//mainBootstrapper.PopulateBaseBootstrappers()
	//
	//err = mainBootstrapper.Bootstrap(serviceCtx)
	//
	//if err != nil {
	//	panic(fmt.Errorf("couldn't bootstrap the node"))
	//}
	//
	//doctorFord, err = newHostManager(cfg, serviceCtx, testAccountMap)
	//
	//if err != nil {
	//	panic(fmt.Errorf("couldn't create new host manager: %w", err))
	//}
	//
	//err = doctorFord.init()
	//
	//if err != nil {
	//	panic(fmt.Errorf("couldn't init host manager: %w", err))
	//}

	//result := m.Run()
	//doctorFord.stop()
	os.Exit(0)
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
