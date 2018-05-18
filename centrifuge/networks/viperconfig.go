package networks

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/viper"
)

const (
	NETWORK_DEFAULT_CONFIG_NAME = "networks"
	NETWORK_DEFAULT_CONFIG_PATH = "../../resources"
	BOOTSTRAP_PEERS_KEY         = "bootstrapPeers"
)

type ViperNetworkConfiguration struct {
	networkString string
	viperConfig   *viper.Viper
}

func (cc *ViperNetworkConfiguration) GetNetworkString() string {
	return cc.networkString
}

func (cc *ViperNetworkConfiguration) GetContractAddress(contract string) (address []byte, err error) {
	address, err = hex.DecodeString(cc.viperConfig.GetString(fmt.Sprintf("contractAddresses.%s", contract))[2:])
	return
}

func (cc *ViperNetworkConfiguration) GetBootstrapPeers() []string {
	return cc.viperConfig.GetStringSlice(BOOTSTRAP_PEERS_KEY)
}

type ViperNetworkConfigurationLoader struct {
	networksConfig    *viper.Viper
	networkConfigName string
	networkConfigPath string
}

func (cl *ViperNetworkConfigurationLoader) SetNetworkConfigName(name string) {
	cl.networkConfigName = name
}

func (cl *ViperNetworkConfigurationLoader) SetNetworkConfigPath(path string) {
	cl.networkConfigPath = path
}

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

func (cl *ViperNetworkConfigurationLoader) GetConfigurationFromKey(key string) (vc *ViperNetworkConfiguration, err error) {
	vc = &ViperNetworkConfiguration{networkString: key}
	vc.viperConfig = cl.networksConfig.Sub(fmt.Sprintf("networks.%s", key))
	return
}

func NewViperNetworkConfigurationLoader() *ViperNetworkConfigurationLoader {
	cl := &ViperNetworkConfigurationLoader{}
	cl.SetNetworkConfigName(NETWORK_DEFAULT_CONFIG_NAME)
	cl.SetNetworkConfigPath(NETWORK_DEFAULT_CONFIG_PATH)
	return cl
}
