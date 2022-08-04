//go:build unit

package config

import (
	"os"
	"testing"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/stretchr/testify/assert"
)

func TestConfiguration_CreateConfigFile(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath()
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	network := "catalyst"
	apiHost := "127.0.0.1"
	apiPort := 8082
	p2pPort := 38202
	bootstrapPeers := []string{
		"/ip4/127.0.0.1/tcp/38202/ipfs/QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk",
		"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
	}
	centChainURL := "ws://127.0.0.1:9946"
	authenticationEnabled := true

	data := map[string]interface{}{
		"targetDataDir":         randomPath,
		"network":               network,
		"bootstraps":            bootstrapPeers,
		"apiHost":               apiHost,
		"apiPort":               apiPort,
		"p2pPort":               p2pPort,
		"p2pConnectTimeout":     "",
		"authenticationEnabled": authenticationEnabled,
		"centChainURL":          centChainURL,
	}

	v, err := CreateConfigFile(data)
	assert.NoError(t, err)

	_, err = os.Stat(randomPath + "/config.yaml")
	assert.NoError(t, err)

	assert.Equal(t, data["network"].(string), v.GetString("centrifugeNetwork"))
	assert.Equal(t, data["bootstraps"].([]string), v.GetStringSlice("networks."+network+".bootstrapPeers"))
	assert.Equal(t, data["apiHost"].(string), v.GetString("nodeHostname"))
	assert.Equal(t, data["apiPort"].(int), v.GetInt("nodePort"))
	assert.Equal(t, data["p2pPort"].(int), v.GetInt("p2p.port"))
	assert.Equal(t, data["p2pConnectTimeout"].(string), v.GetString("p2p.connectTimeout"))
	assert.Equal(t, data["centChainURL"].(string), v.GetString("centChain.nodeURL"))

}

func TestValidateUrl(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		expectedHasErr bool
		expectedURL    string
	}{
		{
			name:           "valid url",
			url:            "http://rinkeby.infura.io/v3",
			expectedHasErr: false,
			expectedURL:    "http://rinkeby.infura.io/v3",
		},
		{
			name:           "without scheme",
			url:            "rinkeby.infura.io/v3/",
			expectedHasErr: false,
			expectedURL:    "https://rinkeby.infura.io/v3/",
		},
		{
			name:           "not allowed scheme",
			url:            "ftp://rinkeby.infura.io/v3/",
			expectedHasErr: true,
			expectedURL:    "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			url, err := validateURL(testCase.url)
			if testCase.expectedHasErr {
				assert.NotNil(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedURL, url)
		})
	}
}
