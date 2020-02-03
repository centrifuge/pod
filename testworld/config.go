// +build testworld

package testworld

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/spf13/viper"
)

// configDir points to the directory containing all configs
const configDir = "configs"

type networkConfig struct {
	// RunNetwork flag is true of you want to spin up a local setup of entire centrifuge network
	RunNetwork bool `json:"run_network"`

	// RunMigrations flag is true when running for first time or network is hard reset
	RunMigrations bool `json:"run_migrations"`

	// CreateHostConfigs flag, make this true this when running in local if configs are not created or outdated
	CreateHostConfigs bool `json:"create_host_configs"`

	Network            string `json:"network"`
	EthNodeURL         string `json:"eth_node_url"`
	TxPoolAccess       bool   `json:"tx_pool_access"`
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

	// if migrations were already run by the wrapper
	if nc.RunMigrations && os.Getenv("MIGRATION_RAN") == "true" {
		log.Info("not running migrations again")
		nc.RunMigrations = false
	}

	// hack to ensure we dont regenerate the configs again.
	dir, err := ioutil.ReadDir(fmt.Sprintf("hostConfigs/%s", nc.Network))
	if err != nil {
		nc.CreateHostConfigs = true
		return nc, nil
	}

	// check if the length of dir is len(hostConfigs)+bernard
	if len(dir) != len(hostConfig)+1 {
		nc.CreateHostConfigs = true
	}

	return nc, nil
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
