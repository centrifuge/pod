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
		"targetDataDir":     targetDir,
		"accountKeyPath":    accountKeyPath,
		"accountPassword":   "pwrd",
		"network":           "russianhill",
		"ethNodeURL":        "http://127.0.0.1:9545",
		"bootstraps":        []string{"/ip4/127.0.0.1/bootstrap1", "/ip4/127.0.0.1/bootstrap2"},
		"apiHost":           "127.0.0.1",
		"apiPort":           int64(8082),
		"p2pPort":           int64(38202),
		"grpcPort":          int64(28202),
		"txpoolaccess":      false,
		"p2pConnectTimeout": "",
		"preCommitEnabled":  false,
		"centChainURL":      "ws://127.0.0.1:9944",
		"centChainID":       "0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
		"centChainSecret":   "0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31",
		"centChainAddr":     "5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR",
	}

	v, err := CreateConfigFile(data)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, data["p2pPort"].(int64), v.GetInt64("p2p.port"), "p2p port match")
	_, err = os.Stat(targetDir + "/config.yaml")
	assert.Nil(t, err, "must be nil, config file should be created")
	c := LoadConfiguration(v.ConfigFileUsed())
	assert.False(t, c.IsPProfEnabled(), "pprof is disabled by default")
	assert.Equal(t, "{}", c.Get("ethereum.accounts.main.key").(string))
	assert.Equal(t, "pwrd", c.Get("ethereum.accounts.main.password").(string))
	bfile, err := ioutil.ReadFile(v.ConfigFileUsed())
	assert.NoError(t, err)
	assert.NotContains(t, string(bfile), "key: \"{}\"")
	assert.NotContains(t, string(bfile), "password: \"pwrd\"")

	cfg := c.(*configuration)
	assert.NotNil(t, cfg.GetP2PResponseDelay())

	assert.NoError(t, os.RemoveAll(targetDir))
}
