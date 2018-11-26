package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../build/configs/...

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/resources"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var log = logging.Logger("config")

// Config defines the methods that a config type should implement.
type Config interface {
	GetStoragePath() string
	GetP2PPort() int
	GetP2PExternalIP() string
	GetP2PConnectionTimeout() time.Duration
	GetReceiveEventNotificationEndpoint() string
	GetServerPort() int
	GetServerAddress() string
	GetNumWorkers() int
	GetWorkerWaitTimeMS() int
	GetEthereumNodeURL() string
	GetEthereumContextReadWaitTimeout() time.Duration
	GetEthereumContextWaitTimeout() time.Duration
	GetEthereumIntervalRetry() time.Duration
	GetEthereumMaxRetries() int
	GetEthereumGasPrice() *big.Int
	GetEthereumGasLimit() uint64
	GetEthereumDefaultAccountName() string
	GetEthereumAccount(accountName string) (account *AccountConfig, err error)
	GetTxPoolAccessEnabled() bool
	GetNetworkString() string
	GetNetworkKey(k string) string
	GetContractAddressString(address string) string
	GetContractAddress(address string) common.Address
	GetBootstrapPeers() []string
	GetNetworkID() uint32
	GetIdentityID() ([]byte, error)
	GetSigningKeyPair() (pub, priv string)
	GetEthAuthKeyPair() (pub, priv string)
}

// Configuration holds the configuration details for the node.
type Configuration struct {
	mu         sync.RWMutex
	configFile string
	v          *viper.Viper
}

// AccountConfig holds the account details.
type AccountConfig struct {
	Address  string
	Key      string
	Password string
}

// IsSet check if the key is set in the config.
func (c *Configuration) IsSet(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.IsSet(key)
}

// Set update the key and the value it holds in the configuration.
func (c *Configuration) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v.Set(key, value)
}

// SetDefault sets the default value for the given key.
func (c *Configuration) SetDefault(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v.SetDefault(key, value)
}

// Get returns associated value for the key.
func (c *Configuration) Get(key string) interface{} {
	return c.get(key)
}

// GetString returns value string associated with key.
func (c *Configuration) GetString(key string) string {
	return cast.ToString(c.get(key))
}

// GetInt returns value int associated with key.
func (c *Configuration) GetInt(key string) int {
	return cast.ToInt(c.get(key))
}

// GetBool returns value bool associated with key.
func (c *Configuration) GetBool(key string) bool {
	return cast.ToBool(c.get(key))
}

// GetDuration returns value duration associated with key.
func (c *Configuration) GetDuration(key string) time.Duration {
	return cast.ToDuration(c.get(key))
}

func (c *Configuration) get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.Get(key)
}

// GetStoragePath returns the data storage backend.
func (c *Configuration) GetStoragePath() string {
	return c.GetString("storage.Path")
}

// GetP2PPort returns P2P Port.
func (c *Configuration) GetP2PPort() int {
	return c.GetInt("p2p.port")
}

// GetP2PExternalIP returns P2P External IP.
func (c *Configuration) GetP2PExternalIP() string {
	return c.GetString("p2p.externalIP")
}

// GetP2PConnectionTimeout returns P2P Connect Timeout.
func (c *Configuration) GetP2PConnectionTimeout() time.Duration {
	return c.GetDuration("p2p.connectTimeout")
}

// GetReceiveEventNotificationEndpoint returns the webhook endpoint defined in the config.
func (c *Configuration) GetReceiveEventNotificationEndpoint() string {
	return c.GetString("notifications.endpoint")
}

// GetServerPort returns the defined server port in the config.
func (c *Configuration) GetServerPort() int {
	return c.GetInt("nodePort")
}

// GetServerAddress returns the defined server address of form host:port in the config.
func (c *Configuration) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.GetString("nodeHostname"), c.GetString("nodePort"))
}

// GetNumWorkers returns number of queue workers defined in the config.
func (c *Configuration) GetNumWorkers() int {
	return c.GetInt("queue.numWorkers")
}

// GetWorkerWaitTimeMS returns the queue worker sleep time between cycles.
func (c *Configuration) GetWorkerWaitTimeMS() int {
	return c.GetInt("queue.workerWaitTimeMS")
}

// GetEthereumNodeURL returns the URL of the Ethereum Node.
func (c *Configuration) GetEthereumNodeURL() string {
	return c.GetString("ethereum.nodeURL")
}

// GetEthereumContextReadWaitTimeout returns the read duration to pass for context.Deadline.
func (c *Configuration) GetEthereumContextReadWaitTimeout() time.Duration {
	return c.GetDuration("ethereum.contextReadWaitTimeout")
}

// GetEthereumContextWaitTimeout returns the commit duration to pass for context.Deadline.
func (c *Configuration) GetEthereumContextWaitTimeout() time.Duration {
	return c.GetDuration("ethereum.contextWaitTimeout")
}

// GetEthereumIntervalRetry returns duration to wait between retries.
func (c *Configuration) GetEthereumIntervalRetry() time.Duration {
	return c.GetDuration("ethereum.intervalRetry")
}

// GetEthereumMaxRetries returns the max acceptable retries.
func (c *Configuration) GetEthereumMaxRetries() int {
	return c.GetInt("ethereum.maxRetries")
}

// GetEthereumGasPrice returns the gas price to use for a ethereum transaction.
func (c *Configuration) GetEthereumGasPrice() *big.Int {
	return big.NewInt(cast.ToInt64(c.get("ethereum.gasPrice")))
}

// GetEthereumGasLimit returns the gas limit to use for a ethereum transaction.
func (c *Configuration) GetEthereumGasLimit() uint64 {
	return cast.ToUint64(c.get("ethereum.gasLimit"))
}

// GetEthereumDefaultAccountName returns the default account to use for the transaction.
func (c *Configuration) GetEthereumDefaultAccountName() string {
	return c.GetString("ethereum.defaultAccountName")
}

// GetEthereumAccount returns the account details associated with the account name.
func (c *Configuration) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	k := fmt.Sprintf("ethereum.accounts.%s", accountName)

	if !c.IsSet(k) {
		return nil, fmt.Errorf("no account found with account name %s", accountName)
	}

	// Workaround for bug https://github.com/spf13/viper/issues/309 && https://github.com/spf13/viper/issues/513
	account = &AccountConfig{
		Address:  c.GetString(fmt.Sprintf("%s.address", k)),
		Key:      c.GetString(fmt.Sprintf("%s.key", k)),
		Password: c.GetString(fmt.Sprintf("%s.password", k)),
	}

	return account, nil
}

// GetTxPoolAccessEnabled returns if the node can check the txpool for nonce increment.
// Note:Important flag for concurrency handling. Disable if Ethereum client doesn't support txpool API (INFURA).
func (c *Configuration) GetTxPoolAccessEnabled() bool {
	return c.GetBool("ethereum.txPoolAccessEnabled")
}

// GetNetworkString returns defined network the node is connected to.
func (c *Configuration) GetNetworkString() string {
	return c.GetString("centrifugeNetwork")
}

// GetNetworkKey returns the specific key(k) value defined in the default network.
func (c *Configuration) GetNetworkKey(k string) string {
	return fmt.Sprintf("networks.%s.%s", c.GetNetworkString(), k)
}

// GetContractAddressString returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddressString(contract string) (address string) {
	return c.GetString(c.GetNetworkKey(fmt.Sprintf("contractAddresses.%s", contract)))
}

// GetContractAddress returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddress(contract string) (address common.Address) {
	return common.HexToAddress(c.GetContractAddressString(contract))
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *Configuration) GetBootstrapPeers() []string {
	return cast.ToStringSlice(c.get(c.GetNetworkKey("bootstrapPeers")))
}

// GetNetworkID returns the numerical network id.
func (c *Configuration) GetNetworkID() uint32 {
	return uint32(c.GetInt(c.GetNetworkKey("id")))
}

// GetIdentityID returns the self centID in bytes.
func (c *Configuration) GetIdentityID() ([]byte, error) {
	id, err := hexutil.Decode(c.GetString("identityId"))
	if err != nil {
		return nil, centerrors.Wrap(err, "can't read identityId from config")
	}
	return id, err
}

// GetSigningKeyPair returns the signing key pair.
func (c *Configuration) GetSigningKeyPair() (pub, priv string) {
	return c.GetString("keys.signing.publicKey"), c.GetString("keys.signing.privateKey")
}

// GetEthAuthKeyPair returns ethereum key pair.
func (c *Configuration) GetEthAuthKeyPair() (pub, priv string) {
	return c.GetString("keys.ethauth.publicKey"), c.GetString("keys.ethauth.privateKey")
}

// IsPProfEnabled returns true if the pprof is enabled
func (c *Configuration) IsPProfEnabled() bool {
	return c.GetBool("debug.pprof")
}

// LoadConfiguration loads the configuration from the given file.
func LoadConfiguration(configFile string) *Configuration {
	cfg := &Configuration{configFile: configFile, mu: sync.RWMutex{}}
	cfg.InitializeViper()
	return cfg
}

func (c *Configuration) readConfigFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	err = c.v.MergeConfig(file)
	return err
}

// InitializeViper loads viper if not loaded already.
// This method should not have any effects if Viper is already initialized.
func (c *Configuration) InitializeViper() {
	if c.v != nil {
		return
	}

	c.v = viper.New()
	c.v.SetConfigType("yaml")

	// Load defaults
	data, err := resources.Asset("go-centrifuge/build/configs/default_config.yaml")
	if err != nil {
		log.Panicf("failed to load (go-centrifuge/build/configs/default_config.yaml): %s", err)
	}

	err = c.v.ReadConfig(bytes.NewReader(data))
	if err != nil {
		log.Panicf("Error reading from default configuration (go-centrifuge/build/configs/default_config.yaml): %s", err)
	}
	// Load user specified config
	if c.configFile != "" {
		log.Infof("Loading user specified config from %s", c.configFile)
		err = c.readConfigFile(c.configFile)
		if err != nil {
			log.Panicf("Error reading config %s, %s", c.configFile, err)
		}
	} else {
		log.Info("No user config specified")
	}
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetEnvPrefix("CENT")
}

// CreateConfigFile creates minimum config file with arguments
func CreateConfigFile(args map[string]interface{}) (*viper.Viper, error) {
	targetDataDir := args["targetDataDir"].(string)
	accountKeyPath := args["accountKeyPath"].(string)
	accountPassword := args["accountPassword"].(string)
	network := args["network"].(string)
	ethNodeURL := args["ethNodeURL"].(string)
	bootstraps := args["bootstraps"].([]string)
	apiPort := args["apiPort"].(int64)
	p2pPort := args["p2pPort"].(int64)
	txPoolAccess := args["txpoolaccess"].(bool)

	if targetDataDir == "" {
		return nil, errors.New("targetDataDir not provided")
	}
	if _, err := os.Stat(targetDataDir); os.IsNotExist(err) {
		os.Mkdir(targetDataDir, os.ModePerm)
	}

	if _, err := os.Stat(accountKeyPath); os.IsNotExist(err) {
		return nil, errors.New("account Key Path does not exist")
	}

	bfile, err := ioutil.ReadFile(accountKeyPath)
	if err != nil {
		return nil, err
	}

	if accountPassword == "" {
		log.Warningf("Account Password not provided")
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("storage.path", targetDataDir+"/db/centrifuge_data.leveldb")
	v.Set("identityId", "")
	v.Set("centrifugeNetwork", network)
	v.Set("nodeHostname", "0.0.0.0")
	v.Set("nodePort", apiPort)
	v.Set("p2p.port", p2pPort)
	v.Set("ethereum.nodeURL", ethNodeURL)
	v.Set("ethereum.txPoolAccessEnabled", txPoolAccess)
	v.Set("ethereum.accounts.main.key", string(bfile))
	v.Set("ethereum.accounts.main.password", accountPassword)
	v.Set("keys.p2p.privateKey", targetDataDir+"/p2p.key.pem")
	v.Set("keys.p2p.publicKey", targetDataDir+"/p2p.pub.pem")
	v.Set("keys.ethauth.privateKey", targetDataDir+"/ethauth.key.pem")
	v.Set("keys.ethauth.publicKey", targetDataDir+"/ethauth.pub.pem")
	v.Set("keys.signing.privateKey", targetDataDir+"/signing.key.pem")
	v.Set("keys.signing.publicKey", targetDataDir+"/signing.pub.pem")

	if bootstraps != nil {
		v.Set("networks."+network+".bootstrapPeers", bootstraps)
	}

	v.SetConfigFile(targetDataDir + "/config.yaml")

	err = v.WriteConfig()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return v, nil
}
