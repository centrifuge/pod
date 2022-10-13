//go:build integration

package config

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
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

		return GenerateP2PKeys(cfg)
	}

	cfg := LoadConfiguration(cfgFile)

	context[bootstrap.BootstrappedConfig] = cfg

	return GenerateP2PKeys(cfg)
}

func (b *Bootstrapper) TestTearDown() error {
	if err := os.RemoveAll(testBootstrapConfigDir); err != nil {
		return fmt.Errorf("couldn't remove temporary config dir: %w", err)
	}

	return nil
}

type CreateTestConfigOpt func(args map[string]any)

func CreateTestConfig(opt CreateTestConfigOpt) (Configuration, string, error) {
	var err error

	testBootstrapConfigDir, err = testingcommons.GetRandomTestStoragePath("config-test-bootstrapper-*")

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create temp dir: %w", err)
	}

	args := map[string]any{
		"targetDataDir": testBootstrapConfigDir,
		"network":       "test",
		"bootstraps":    []string{},
		"apiPort":       getRandomPort(37000, 38000),
		"p2pPort":       getRandomPort(38000, 39000),
		// TODO(cdamian): Lower this timeout when done with debugging.
		"p2pConnectTimeout":      "5m",
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

	if opt != nil {
		opt(args)
	}

	cfgFile, err := CreateConfigFile(args)

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create config file: %w", err)
	}

	cfg := LoadConfiguration(cfgFile.ConfigFileUsed())

	return cfg, cfgFile.ConfigFileUsed(), nil
}

func getRandomPort(min, max int) int {
	p := rand.Intn(max - min)
	return p + min
}
