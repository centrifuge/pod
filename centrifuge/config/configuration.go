package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -ignore=data\\.go -o ../resources/data.go ../../resources/...

import (
	"bytes"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/resources"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	NETWORK_DEFAULT_CONFIG_FILE = "resources/networks.yaml"
)

var log = logging.Logger("config")
var Config Configuration

type Configuration struct {
	configFile string
	V          *viper.Viper
}

// Storage backend
func (c *Configuration) GetStoragePath() string {
	return c.V.GetString("storage.Path")
}

// P2P
func (c *Configuration) GetP2PPort() int {
	return c.V.GetInt("p2p.port")
}

////////////////////////////////////////////////////////////////////////////////
// Network Configuration
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetNetworkString() string {
	return c.V.GetString("centrifugeNetwork")
}

func (c *Configuration) GetNetworkConfig() *viper.Viper {
	key := fmt.Sprintf("networks.%s", c.GetNetworkString())
	if !c.V.IsSet(key) {
		log.Panicf("networkConfig: Network configuration with key [%] does not exist", key)
	}
	return c.V.Sub(key)
}

// GetContractAddress returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddress(contract string) (address string) {
	return c.GetNetworkConfig().GetString(fmt.Sprintf("contractAddresses.%s", contract))
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *Configuration) GetBootstrapPeers() []string {
	return c.GetNetworkConfig().GetStringSlice("bootstrapPeers")
}

// GetNetworkID returns the numerical network id.
func (c *Configuration) GetNetworkID() int64 {
	return c.GetNetworkConfig().GetInt64("id")
}

// SigningKeys
func (c *Configuration) GetSigningKeyPair() (pub, priv string) {
	return c.V.GetString("keys.signing.publicKey"), c.V.GetString("keys.signing.privateKey")
}

// Configuration Implementation
func NewConfiguration(configFile string) Configuration {
	c := Configuration{configFile: configFile}
	return c
}

func (c *Configuration) ReadConfigFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	err = c.V.MergeConfig(file)
	return err
}

func (c *Configuration) InitializeViper() {
	c.V = viper.New()
	c.V.SetConfigType("yaml")

	// Load defaults
	data, _ := resources.Asset(NETWORK_DEFAULT_CONFIG_FILE)
	err := c.V.ReadConfig(bytes.NewReader(data))
	if err != nil {
		log.Panicf("Error reading config %s, %s", NETWORK_DEFAULT_CONFIG_FILE, err)
	}

	// Load user specified config
	err = c.ReadConfigFile(c.configFile)
	if err != nil {
		log.Panicf("Error reading config %s, %s", c.configFile, err)
	}
	c.V.AutomaticEnv()
	c.V.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.V.SetEnvPrefix("CENT")
}

func Bootstrap(configFile string) {
	Config = NewConfiguration(configFile)
	Config.InitializeViper()
}
