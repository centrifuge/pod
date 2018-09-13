package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../../resources/...

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/resources"
	"github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/viper"
)

var log = logging.Logger("config")
var Config *Configuration

type Configuration struct {
	configFile string
	V          *viper.Viper
}

type AccountConfig struct {
	Address  string
	Key      string
	Password string
}

// IdentityConfig holds ID, public and private key of a single entity
type IdentityConfig struct {
	ID         []byte
	PublicKey  []byte
	PrivateKey []byte
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
// Notifications
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetReceiveEventNotificationEndpoint() string {
	return c.V.GetString("notifications.endpoint")
}

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
// Queuing
////////////////////////////////////////////////////////////////////////////////

func (c *Configuration) GetNumWorkers() int {
	return c.V.GetInt("queue.numWorkers")
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

func (c *Configuration) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	k := fmt.Sprintf("ethereum.accounts.%s", accountName)

	if !c.V.IsSet(k) {
		return nil, fmt.Errorf("No account found with account name %s", accountName)
	}

	// Workaround for bug https://github.com/spf13/viper/issues/309 && https://github.com/spf13/viper/issues/513
	account = &AccountConfig{
		Address:  c.V.GetString(fmt.Sprintf("%s.address", k)),
		Key:      c.V.GetString(fmt.Sprintf("%s.key", k)),
		Password: c.V.GetString(fmt.Sprintf("%s.password", k)),
	}

	return account, nil
}

// Important flag for concurrency handling. Disable if Ethereum client doesn't support txpool API (INFURA)
func (c *Configuration) GetTxPoolAccessEnabled() bool {
	return c.V.GetBool("ethereum.txPoolAccessEnabled")
}

////////////////////////////////////////////////////////////////////////////////
// Network Configuration
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetNetworkString() string {
	return c.V.GetString("centrifugeNetwork")
}

func (c *Configuration) GetNetworkKey(k string) string {
	return fmt.Sprintf("networks.%s.%s", c.GetNetworkString(), k)
}

// GetContractAddressString returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddressString(contract string) (address string) {
	return c.V.GetString(c.GetNetworkKey(fmt.Sprintf("contractAddresses.%s", contract)))
}

// GetContractAddress returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddress(contract string) (address common.Address) {
	return common.HexToAddress(c.GetContractAddressString(contract))
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *Configuration) GetBootstrapPeers() []string {
	return c.V.GetStringSlice(c.GetNetworkKey("bootstrapPeers"))
}

// GetNetworkID returns the numerical network id.
func (c *Configuration) GetNetworkID() uint32 {
	return uint32(c.V.GetInt(c.GetNetworkKey("id")))
}

// Identity:
func (c *Configuration) GetIdentityId() []byte {
	return []byte(c.V.GetString("identityId"))
}

func (c *Configuration) GetSigningKeyPair() (pub, priv string) {
	return c.V.GetString("keys.signing.publicKey"), c.V.GetString("keys.signing.privateKey")
}

func (c *Configuration) GetEthAuthKeyPair() (pub, priv string) {
	return c.V.GetString("keys.ethauth.publicKey"), c.V.GetString("keys.ethauth.privateKey")
}

// Configuration Implementation
func NewConfiguration(configFile string) *Configuration {
	c := Configuration{configFile: configFile}
	return &c
}

// SetConfigFile returns an error if viper was already initialized.
func (c *Configuration) SetConfigFile(path string) error {
	if c.V != nil {
		return errors.New("Viper already initialized. Can't set config file")
	}
	c.configFile = path
	return nil
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
