//go:build integration

package config

import (
	"fmt"
	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
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
		var err error

		cfgFile, err = generateRandomConfig()

		if err != nil {
			return err
		}
	}

	context[BootstrappedConfigFile] = cfgFile

	cfg := LoadConfiguration(cfgFile)

	if err := GenerateNodeKeys(cfg); err != nil {
		return fmt.Errorf("couldn't generate node keys: %w", err)
	}

	context[bootstrap.BootstrappedConfig] = cfg

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	if err := os.RemoveAll(testBootstrapConfigDir); err != nil {
		return fmt.Errorf("couldn't remove temporary config dir: %w", err)
	}

	return nil
}

func generateRandomConfig() (string, error) {
	var err error

	testBootstrapConfigDir, err = testingcommons.GetRandomTestStoragePath("config-test-bootstrapper")

	if err != nil {
		return "", fmt.Errorf("couldn't create temp dir: %w", err)
	}

	args := map[string]any{
		"targetDataDir": testBootstrapConfigDir,
		"network":       "test",
		"bootstraps": []string{
			"/ip4/127.0.0.1/tcp/38202/ipfs/QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk",
			"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
		},
		"apiPort":                8082,
		"p2pPort":                38202,
		"p2pConnectTimeout":      "",
		"apiHost":                "127.0.0.1",
		"authenticationEnabled":  true,
		"ipfsPinningServiceName": "pinata",
		"ipfsPinningServiceURL":  "https://pinata.com",
		"ipfsPinningServiceAuth": "test-auth",
		// Eve's secret seed
		"podAdminSecretSeed": "0x786ad0e2df456fe43dd1f91ebca22e235bc162e0bb8d53c633e8c85b2af68b7a",
		// Ferdie's secret seed
		"podOperatorSecretSeed": "0x42438b7883391c05512a938e36c2df0131e088b3756d6aa7a755fbff19d2f842",
		"centChainURL":          "ws://127.0.0.1:9946",
	}

	cfg, err := CreateConfigFile(args)

	if err != nil {
		return "", fmt.Errorf("couldn't create config file: %w", err)
	}

	return cfg.ConfigFileUsed(), nil
}
