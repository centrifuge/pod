// +build testworld

package testworld

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// configFile is the main configuration file, used mainly on CI
const configFile = "configs/base.json"

// localConfigFile is the local configuration file, when running locally for the first time,
// copy the configFile above to this location so that you can customise for local.
const localConfigFile = "configs/local/local.json"

type testConfig struct {

	// following are env specific configs

	// OtherLocalConfig is used to direct running of Testworld with different config than local json, eg: for switching networks for local testing.
	// Note that if you change this all other config fields in localConfigFile is ignored.
	OtherLocalConfig string `json:"otherLocalConfig"`

	// RunChains flag, Adjust this based on local testing requirements in localConfigFile
	RunChains bool `json:"runChains"`

	// CreateHostConfigs flag, make this true this when running for the first time in local env using localConfigFile
	CreateHostConfigs bool `json:"createHostConfigs"`

	// RunMigrations flag, make this false if you want to make the tests run faster locally in localConfigFile
	RunMigrations bool `json:"runMigrations"`

	// following are host(cent node) specific configs

	EthNodeURL      string `json:"ethNodeURL"`
	AccountKeyPath  string `json:"accountKeyPath"`
	AccountPassword string `json:"accountPassword"`
	Network         string `json:"network"`
	TxPoolAccess    bool   `json:"txPoolAccess"`
}

func loadConfig(isLocal bool) (testConfig, string, error) {
	c := configFile
	if isLocal {
		c = localConfigFile
	} else {
		fmt.Printf("Testworld using config %s\n", configFile)
	}
	var config testConfig
	conf, err := os.Open(c)
	if err != nil {
		if isLocal {
			// load the base config if the local config is not available
			return loadConfig(false)
		} else {
			fmt.Println(err.Error())
			return config, "", err
		}
	}
	defer conf.Close()
	jsonParser := json.NewDecoder(conf)
	err = jsonParser.Decode(&config)
	if err != nil {
		fmt.Println(err.Error())
		return config, "", err
	}

	// load custom config specified in OtherLocalConfig, if this is local
	if isLocal && config.OtherLocalConfig != "" {
		customConfig, err := loadCustomLocalConfig(config.OtherLocalConfig)
		if err != nil {
			// use the default local config
			fmt.Printf("Testworld using config %s\n", localConfigFile)
			return config, extractConfigName(localConfigFile), nil
		}
		fmt.Printf("Testworld using config %s\n", config.OtherLocalConfig)
		return customConfig, extractConfigName(config.OtherLocalConfig), nil
	} else if isLocal {
		// using the default local config
		fmt.Printf("Testworld using config %s\n", localConfigFile)
		return config, extractConfigName(localConfigFile), nil
	}
	return config, extractConfigName(configFile), nil
}

func loadCustomLocalConfig(customConfigFile string) (testConfig, error) {
	var config testConfig
	conf, err := os.Open(customConfigFile)
	if err != nil {
		fmt.Println(err.Error())
		return config, err
	}
	defer conf.Close()
	jsonParser := json.NewDecoder(conf)
	err = jsonParser.Decode(&config)
	if err != nil {
		fmt.Println(err.Error())
		return config, err
	}
	return config, nil
}

func extractConfigName(path string) string {
	filename := filepath.Base(path)
	parts := strings.Split(filename, ".")
	return parts[0]
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
