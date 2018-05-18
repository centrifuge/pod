package networks

import (
	"github.com/spf13/viper"
)

const (
	NETWORK_DEFAULT_CONFIG_NAME = "networks"
	NETWORK_DEFAULT_CONFIG_PATH = "resources"
)

type ViperNetworkConfiguration struct {
	networkString string
	viperConfig   *viper.Viper
}

func (cc *ViperNetworkConfiguration) GetNetworkString() string {
	return cc.networkString
}

func (cc *ViperNetworkConfiguration) GetContractAddress(string) []byte {
	return []byte{}
}

func (cc *ViperNetworkConfiguration) GetBootstrapPeers(string) []string {
	return []string{""}
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
	c.SetConfigName("networks")
	c.AddConfigPath("./resources")
	err = c.ReadInConfig()
	if err != nil {
		return err
	}
	cl.networksConfig = c
	return
}

func (cl *ViperNetworkConfigurationLoader) GetConfigurationFromKey(key string) (vc *ViperNetworkConfiguration, err error) {
	vc = &ViperNetworkConfiguration{networkString: key}
	vc.viperConfig = cl.networksConfig.Sub(key)
	return
}

func NewViperNetworkConfigurationLoader() *ViperNetworkConfigurationLoader {
	cl := &ViperNetworkConfigurationLoader{}
	cl.SetNetworkConfigName(NETWORK_DEFAULT_CONFIG_NAME)
	cl.SetNetworkConfigPath(NETWORK_DEFAULT_CONFIG_PATH)
	return cl
}
