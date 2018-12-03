package testworld

import (
	"encoding/json"
	"fmt"
	"os"
)

type testConfig struct {
	EthNodeUrl      string `json:"ethNodeUrl"`
	AccountKeyPath  string `json:"accountKeyPath"`
	AccountPassword string `json:"accountPassword"`
	Network         string `json:"network"`
	TxPoolAccess    bool   `json:"txPoolAccess"`
}

func loadConfig(file string) (testConfig, error) {
	var config testConfig
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, nil
}
