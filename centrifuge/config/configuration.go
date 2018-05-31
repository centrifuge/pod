package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../../resources/...

import (
	"bytes"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/resources"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/viper"
	"math/big"
	"os"
	"strings"
	"time"
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

//

////////////////////////////////////////////////////////////////////////////////
// Server
////////////////////////////////////////////////////////////////////////////////

func (c *Configuration) GetServerPort() int {
	return c.V.GetInt("nodePort")
}

func (c *Configuration) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.V.GetString("nodeHostname"), c.V.GetString("nodePort"))
}

////////////////////////////////////////////////////////////////////////////////
// Ethereum
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetEthereumNodeURL() string {
	return c.V.GetString("ethereum.nodeURL")
}

func (c *Configuration) GetEthereumContextWaitTimeout() time.Duration {
	return c.V.GetDuration("ethereum.contextWaitTimeout")
}

func (c *Configuration) GetEthereumIntervalRetry() time.Duration {
	return c.V.GetDuration("ethereum.intervalRetry")
}

func (c *Configuration) GetEthereumMaxRetries() int {
	return c.V.GetInt("ethereum.maxRetries")
}

func (c *Configuration) GetEthereumGasPrice() *big.Int {
	return big.NewInt(c.V.GetInt64("ethereum.gasPrice"))
}

func (c *Configuration) GetEthereumGasLimit() uint64 {
	return uint64(c.V.GetInt64("ethereum.gasLimit"))
}

func (c *Configuration) GetEthereumDefaultAccountName() string {
	return c.V.GetString("ethereum.defaultAccountName")
}

func (c *Configuration) GetEthereumAccountMap(accountName string) (accounts map[string]string, err error) {
	k := fmt.Sprintf("ethereum.accounts.%s", accountName)

	if !c.V.IsSet(k) {
		return nil, fmt.Errorf("No account found with account name %s", accountName)
	}
	return c.V.GetStringMapString(k), nil
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
func (c *Configuration) GetNetworkID() uint32 {
	return uint32(c.GetNetworkConfig().GetInt("id"))
}

// Identity:
func (c *Configuration) GetIdentityId() []byte {
	return []byte(c.V.GetString("identityId"))
}

// GetKnownSigningKeys is just a hack until we have implemented
func (c *Configuration) GetKnownSigningKeys() map[string]string {
	return c.V.GetStringMapString("keys.knownSigningKeys")
}

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
	// This method should not have any effects if Viper is already initialized.
	if c.V != nil {
		return
	}

	c.V = viper.New()
	c.V.SetConfigType("yaml")

	// Load defaults
	data, _ := resources.Asset("resources/default_config.yaml")
	err := c.V.ReadConfig(bytes.NewReader(data))
	if err != nil {
		log.Panicf("Error reading from default configuration (resources/default_config.yaml): %s", err)
	}

	// Load user specified config
	if c.configFile != "" {
		log.Infof("Loading user specified config from %s", c.configFile)
		err = c.ReadConfigFile(c.configFile)
		if err != nil {
			log.Panicf("Error reading config %s, %s", c.configFile, err)
		}
	} else {
		log.Info("No user config specified")
	}
	c.V.AutomaticEnv()
	c.V.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.V.SetEnvPrefix("CENT")
}

func Bootstrap(configFile string) {
	Config = NewConfiguration(configFile)
	Config.InitializeViper()
}
