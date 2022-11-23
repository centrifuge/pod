//go:build integration || testworld

package config

import (
	"fmt"
	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/crypto"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
)

var (
	testBootstrapConfigDir string
)

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; ok {
		return nil
	}

	cfgFile, ok := context[BootstrappedConfigFile].(string)

	if !ok {
		cfg, cfgFile, err := CreateTestConfig(nil)

		if err != nil {
			return err
		}

		context[BootstrappedConfigFile] = cfgFile
		context[bootstrap.BootstrappedConfig] = cfg

		return GenerateAndWriteP2PKeys(cfg)
	}

	cfg := LoadConfiguration(cfgFile)

	context[bootstrap.BootstrappedConfig] = cfg

	return GenerateAndWriteP2PKeys(cfg)
}

func (b *Bootstrapper) TestTearDown() error {
	if err := os.RemoveAll(testBootstrapConfigDir); err != nil {
		return fmt.Errorf("couldn't remove temporary config dir: %w", err)
	}

	return nil
}

type ArgsOverrideFn func(cfgArgs map[string]any)

func CreateTestConfig(argsOverrideFn ArgsOverrideFn) (Configuration, string, error) {
	var err error

	testBootstrapConfigDir, err = testingcommons.GetRandomTestStoragePath("config-test-bootstrapper-*")

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create temp dir: %w", err)
	}

	podOperatorSecretSeed, err := crypto.GenerateSR25519SecretSeed()

	if err != nil {
		return nil, "", fmt.Errorf("couldn't generate pod operator secret seed: %w", err)
	}

	args := map[string]any{
		"targetDataDir":          testBootstrapConfigDir,
		"network":                "test",
		"bootstraps":             []string{},
		"apiPort":                testingcommons.MustGetFreePort(),
		"p2pPort":                testingcommons.MustGetFreePort(),
		"p2pConnectTimeout":      "1m",
		"apiHost":                "127.0.0.1",
		"authenticationEnabled":  true,
		"ipfsPinningServiceName": "pinata",
		"ipfsPinningServiceURL":  "https://pinata.com",
		"ipfsPinningServiceAuth": "test-auth",
		// Eve's secret seed
		"podAdminSecretSeed":    "0x786ad0e2df456fe43dd1f91ebca22e235bc162e0bb8d53c633e8c85b2af68b7a",
		"podOperatorSecretSeed": podOperatorSecretSeed,
		"centChainURL":          "ws://127.0.0.1:9946",
	}

	if argsOverrideFn != nil {
		argsOverrideFn(args)
	}

	cfgFile, err := CreateConfigFile(args)

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create config file: %w", err)
	}

	cfg := LoadConfiguration(cfgFile.ConfigFileUsed())

	return cfg, cfgFile.ConfigFileUsed(), nil
}
