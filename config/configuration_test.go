// +build unit

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfiguration_CreateConfigFile(t *testing.T) {
	targetDir := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	accountKeyPath := targetDir + "/main.key"
	err := os.Mkdir(targetDir, os.ModePerm)
	assert.Nil(t, err, "err should be nil")

	err = ioutil.WriteFile(accountKeyPath, []byte("{}"), os.ModePerm)
	assert.Nil(t, err, "err should be nil")

	data := map[string]interface{}{
		"targetDataDir":   targetDir,
		"accountKeyPath":  accountKeyPath,
		"accountPassword": "pwrd",
		"network":         "russianhill",
		"ethNodeURL":      "ws://127.0.0.1:9546",
		"bootstraps":      []string{"/ip4/127.0.0.1/bootstrap1", "/ip4/127.0.0.1/bootstrap2"},
		"apiPort":         int64(8082),
		"p2pPort":         int64(38202),
		"txpoolaccess":    false,
	}

	v, err := CreateConfigFile(data)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, data["p2pPort"].(int64), v.GetInt64("p2p.port"), "p2p port match")
	_, err = os.Stat(targetDir + "/config.yaml")
	assert.Nil(t, err, "must be nil, config file should be created")
	c := LoadConfiguration(v.ConfigFileUsed())
	assert.False(t, c.IsPProfEnabled(), "pprof is disabled by default")
	os.Remove(targetDir)
}
