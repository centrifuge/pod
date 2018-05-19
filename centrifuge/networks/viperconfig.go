package networks

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

const (
	NETWORK_DEFAULT_CONFIG_NAME = "networks"
	NETWORK_DEFAULT_CONFIG_PATH = "../../resources"
	BOOTSTRAP_PEERS_KEY         = "bootstrapPeers"
)

// ViperNetworkConfiguration is a NetworkConfiguration implementation that uses
// Viper to fetch configs from a compatible source.
type ViperNetworkConfiguration struct {
	networkString string
	viperConfig   *viper.Viper
}

// GetNetworkString returns the unique network identifier
func (cc *ViperNetworkConfiguration) GetNetworkString() string {
	return cc.networkString
}

// GetContractAddress returns the deployed contract address for a given contract.
func (cc *ViperNetworkConfiguration) GetContractAddress(contract string) (address []byte, err error) {
	address, err = hex.DecodeString(
		cc.viperConfig.GetString(fmt.Sprintf("contractAddresses.%s", contract))[2:])
	return
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (cc *ViperNetworkConfiguration) GetBootstrapPeers() []string {
	return cc.viperConfig.GetStringSlice(BOOTSTRAP_PEERS_KEY)
}

// ViperNetworkConfigurationLoader loads a configuration for a network by it's network key
// from a config file
type ViperNetworkConfigurationLoader struct {
	networksConfig    *viper.Viper
	networkConfigName string
	networkConfigPath string
}

// SetNetworkConfigName sets the configuration file name to look for
func (cl *ViperNetworkConfigurationLoader) SetNetworkConfigName(name string) {
	cl.networkConfigName = name
}

// SetNetworkConfigName sets the path to look search for the file in
func (cl *ViperNetworkConfigurationLoader) SetNetworkConfigPath(path string) {
	cl.networkConfigPath = path
}

// LoadNetworkConfig loads the config file by default located in resources/networks.yaml
func (cl *ViperNetworkConfigurationLoader) LoadNetworkConfig() (err error) {
	c := viper.New()
	c.SetConfigName(cl.networkConfigName)
	c.AddConfigPath(cl.networkConfigPath)
	err = c.ReadInConfig()
	if err != nil {
		return err
	}
	cl.networksConfig = c
	return
}

// GetConfigurationFromKey gets the network specific configuration from the list
// of networks defined in the networks configuration file
func (cl *ViperNetworkConfigurationLoader) GetConfigurationFromKey(key string) (vc *ViperNetworkConfiguration, err error) {
	vc = &ViperNetworkConfiguration{networkString: key}
	key = fmt.Sprintf("networks.%s", key)
	if !cl.networksConfig.IsSet(key) {
		return nil, errors.New("networkConfig: Network configuration does not exist")
	}
	vc.viperConfig = cl.networksConfig.Sub(key)
	return
}

// NewViperNetworkConfigurationLoader returns a ViperNetworkConfigurationLoader configured
// with the default options `NETWORK_DEFUALT_CONFIG_NAME` and `NETWORK_DEFAULT_CONFIG_PATH`
func NewViperNetworkConfigurationLoader() *ViperNetworkConfigurationLoader {
	cl := &ViperNetworkConfigurationLoader{}
	cl.SetNetworkConfigName(NETWORK_DEFAULT_CONFIG_NAME)
	cl.SetNetworkConfigPath(NETWORK_DEFAULT_CONFIG_PATH)
	return cl
}
