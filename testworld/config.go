// +build testworld

package testworld

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"os"
	"strings"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/spf13/viper"
)

// configDir points to the directory containing all configs
const configDir = "configs"

const ethNodeURL = "ETH_NODE_URL"

type networkConfig struct {
	// RunNetwork flag is true of you want to spin up a local setup of entire centrifuge network
	RunNetwork bool `json:"run_network"`

	// RunMigrations flag is true when running for first time or network is hard reset
	RunMigrations bool `json:"run_migrations"`

	// CreateHostConfigs flag, make this true this when running in local if configs are not created or outdated
	CreateHostConfigs bool `json:"create_host_configs"`

	Network            string `json:"network"`
	EthNodeURL         string `json:"eth_node_url"`
	EthAccountKeyPath  string `json:"eth_account_key_path"`
	EthAccountPassword string `json:"eth_account_password"`

	CentChainURL        string `json:"cent_chain_url"`
	CentChainSecret     string `json:"cent_chain_secret"`
	CentChainAccountID  string `json:"cent_chain_account_id"`
	CentChainS58Address string `json:"cent_chain_s58_address"`

	ContractAddresses *config.SmartContractAddresses `json:"contract_addresses"`
	DappAddresses     map[string]string              `json:"dapp_addresses"`
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

	if nc.EthNodeURL == "" {
		url := getEnv(ethNodeURL, false)
		if url == "" {
			return nc, errors.New("Eth node URL is empty")
		}
		nc.EthNodeURL = url
	}

	err = checkEthKeyPath(nc.Network, nc.EthAccountKeyPath)
	if err != nil {
		return nc, err
	}

	if nc.EthAccountPassword == "" {
		nc.EthAccountPassword = getEnv("ETH_"+nc.Network+"_SECRET", true)
	}

	if nc.CentChainSecret == "" {
		nc.CentChainSecret = getEnv("CC_"+nc.Network+"_SECRET", false)
	}

	// if migrations were already run by the wrapper
	if nc.RunMigrations && os.Getenv("MIGRATION_RAN") == "true" {
		log.Info("not running migrations again")
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

func checkEthKeyPath(network, file string) error {
	if _, err := os.Stat(file); err == nil {
		return nil
	}

	val := getEnv("ETH_"+network+"_KEY", true)
	if val == "" {
		return errors.New("failed to get eth key from the env")
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync()
	_, err = f.WriteString(val)
	return err
}
