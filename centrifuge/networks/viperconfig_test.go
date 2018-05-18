// +build unit

package networks

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestViperNetworkConfigurationLoader_LoadNetworkConfig(t *testing.T) {
	cl := NewViperNetworkConfigurationLoader()
	err := cl.LoadNetworkConfig()
	assert.Nil(t, err)
	// Check a known value from the default configuration
	assert.Equal(t, 4, cl.networksConfig.GetInt("networks.centrifuge-russianhill-eth-rinkeby.ethereumNetworkId"))

	networkString := "centrifuge-russianhill-eth-rinkeby"
	conf, err := cl.GetConfigurationFromKey(networkString)
	assert.Nil(t, err)
	assert.Equal(t, networkString, conf.GetNetworkString())

	contractId, err := conf.GetContractAddress("identityFactory")
	fmt.Println(err)
	expectedContractId, _ := hex.DecodeString("0589ed482af8d6809f022fc11aa399fc8a883d52")
	assert.Equal(t, expectedContractId, contractId)

	expectedBootstrapPeers := []string{
		"/ip4/127.0.0.1/tcp/38202/ipfs/QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk",
		"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
	}
	bootstrapPeers := conf.GetBootstrapPeers()
	assert.Equal(t, expectedBootstrapPeers, bootstrapPeers)

	// Try to load a nonexistent configuration
	cl = &ViperNetworkConfigurationLoader{
		networkConfigPath: ".",
		networkConfigName: "doesnotexist",
	}
	err = cl.LoadNetworkConfig()
	assert.Error(t, err)
	fmt.Println(err)

}
