//go:build testworld

package testworld

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

const (
	testDir             = "go-centrifuge-test"
	testworldDirPattern = "testworld-*"

	defaultTestworldConfigNetwork               = "testing"
	defaultTestworldConfigAPIPort               = 8082
	defaultTestworldConfigP2PPort               = 38204
	defaultTestworldConfigAPIHost               = "127.0.0.1"
	defaultTestworldConfigAuthenticationEnabled = true
	defaultTestworldConfigCentChainURL          = "ws://localhost:9946"
	defaultTestworldIPFSPinningServiceName      = "pinata"
	defaultTestworldIPFSPinningServiceURL       = "https://api.pinata.cloud"
	defaultTestworldIPFSPinningServiceAuth      = ""

	// Ferdie's secret seed
	defaultTestworldPodOperatorSecretSeed = "0x42438b7883391c05512a938e36c2df0131e088b3756d6aa7a755fbff19d2f842"

	// Eve's secret seed
	defaultTestworldPodAdminSecretSeed = "0x786ad0e2df456fe43dd1f91ebca22e235bc162e0bb8d53c633e8c85b2af68b7a"
)

func getConfigVals() (map[string]any, error) {
	tempDir := path.Join(os.TempDir(), testDir)
	err := os.MkdirAll(tempDir, os.ModePerm)

	if err != nil {
		return nil, err
	}

	dirPath, err := os.MkdirTemp(tempDir, testworldDirPattern)

	if err != nil {
		return nil, err
	}

	var bootstrapPeers []string

	return map[string]any{
		"targetDataDir":          dirPath,
		"network":                defaultTestworldConfigNetwork,
		"bootstraps":             bootstrapPeers,
		"apiPort":                defaultTestworldConfigAPIPort,
		"p2pPort":                defaultTestworldConfigP2PPort,
		"p2pConnectTimeout":      "",
		"apiHost":                defaultTestworldConfigAPIHost,
		"authenticationEnabled":  defaultTestworldConfigAuthenticationEnabled,
		"centChainURL":           defaultTestworldConfigCentChainURL,
		"ipfsPinningServiceName": defaultTestworldIPFSPinningServiceName,
		"ipfsPinningServiceURL":  defaultTestworldIPFSPinningServiceURL,
		"ipfsPinningServiceAuth": defaultTestworldIPFSPinningServiceAuth,
		"podOperatorSecretSeed":  defaultTestworldPodOperatorSecretSeed,
		"podAdminSecretSeed":     defaultTestworldPodAdminSecretSeed,
	}, nil
}

// configDir points to the directory containing all configs
const configDir = "configs"

type networkConfig struct {
	// RunNetwork flag is true of you want to spin up a local setup of entire centrifuge network
	RunNetwork bool `json:"run_network"`

	// RunMigrations flag is true when running for first time or network is hard reset
	RunMigrations bool `json:"run_migrations"`

	// CreateHostConfigs flag, make this true this when running in local if configs are not created or outdated
	CreateHostConfigs bool `json:"create_host_configs"`

	Network string `json:"network"`

	CentChainURL string `json:"cent_chain_url"`
}

func loadConfig(network string) (nc networkConfig, err error) {
	file := fmt.Sprintf("%s/%s.json", configDir, network)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nc, err
	}

	err = json.Unmarshal(data, &nc)
	if err != nil {
		return nc, err
	}

	// if migrations were already run by the wrapper
	if nc.RunMigrations && os.Getenv("MIGRATION_RAN") == "true" {
		//log.Info("not running migrations again")
		nc.RunMigrations = false
	}

	// hack to ensure we dont regenerate the configs again.
	_, err = ioutil.ReadDir(fmt.Sprintf("hostConfigs/%s", nc.Network))
	if err != nil {
		nc.CreateHostConfigs = true
		return nc, nil
	}

	// ensure sleepy config is deleted as it will be created later
	return nc, os.RemoveAll(fmt.Sprintf("hostConfigs/%s/Sleepy", nc.Network))
}

func updateConfig(dir string, values map[string]interface{}) (err error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(dir + "/config.yaml")
	err = v.ReadInConfig()
	if err != nil {
		return err
	}

	for k, val := range values {
		v.Set(k, val)
	}

	return v.WriteConfig()
}

// decodes the env and returns the result
func getEnv(env string, encoded bool) string {
	v := strings.TrimSpace(os.Getenv(env))
	if v == "" {
		return v
	}
	if encoded {
		d, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return ""
		}
		v = strings.TrimSpace(string(d))
	}
	return v
}
