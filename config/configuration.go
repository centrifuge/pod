package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../build/configs/...

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/centrifuge/go-centrifuge/resources"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var log = logging.Logger("config")

// TenantKey is used as key for the tenant identity in the context.ContextWithValue.
var TenantKey struct{}

// ContractName is a type to indicate a contract name parameter
type ContractName string

const (
	// AnchorRepo is the contract name for AnchorRepo
	AnchorRepo ContractName = "anchorRepository"

	// IdentityFactory is the contract name for IdentityFactory
	IdentityFactory ContractName = "identityFactory"

	// IdentityRegistry is the contract name for IdentityRegistry
	IdentityRegistry ContractName = "identityRegistry"

	// PaymentObligation is the contract name for PaymentObligation
	PaymentObligation ContractName = "paymentObligation"
)

// ContractNames returns the list of smart contract names currently used in the system, please update this when adding new contracts
func ContractNames() [4]ContractName {
	return [4]ContractName{AnchorRepo, IdentityFactory, IdentityRegistry, PaymentObligation}
}

// Configuration defines the methods that a config type should implement.
type Configuration interface {
	storage.Model

	// generic methods
	IsSet(key string) bool
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
	SetupSmartContractAddresses(network string, smartContractAddresses *SmartContractAddresses)
	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetDuration(key string) time.Duration

	GetStoragePath() string
	GetConfigStoragePath() string
	GetTenantsKeystore() string
	GetP2PPort() int
	GetP2PExternalIP() string
	GetP2PConnectionTimeout() time.Duration
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
	GetTxPoolAccessEnabled() bool
	GetNetworkString() string
	GetNetworkKey(k string) string
	GetContractAddressString(address string) string
	GetContractAddress(contractName ContractName) common.Address
	GetBootstrapPeers() []string
	GetNetworkID() uint32

	// CentID specific configs (eg: for multi tenancy)
	GetEthereumAccount(accountName string) (account *AccountConfig, err error)
	GetEthereumDefaultAccountName() string
	GetReceiveEventNotificationEndpoint() string
	GetIdentityID() ([]byte, error)
	GetSigningKeyPair() (pub, priv string)
	GetEthAuthKeyPair() (pub, priv string)

	// debug specific methods
	IsPProfEnabled() bool

	// CreateProtobuf creates protobuf
	CreateProtobuf() *configpb.ConfigData
}

// TenantConfiguration exposes tenant specific config options
type TenantConfiguration interface {
	storage.Model

	GetEthereumAccount() *AccountConfig
	GetEthereumDefaultAccountName() string
	GetReceiveEventNotificationEndpoint() string
	GetIdentityID() ([]byte, error)
	GetSigningKeyPair() (pub, priv string)
	GetEthAuthKeyPair() (pub, priv string)
	GetEthereumContextWaitTimeout() time.Duration

	// CreateProtobuf creates protobuf
	CreateProtobuf() *configpb.TenantData
}

// Service exposes functions over the config objects
type Service interface {
	GetConfig() (Configuration, error)
	GetTenant(identifier []byte) (TenantConfiguration, error)
	GetAllTenants() ([]TenantConfiguration, error)
	CreateConfig(data Configuration) (Configuration, error)
	CreateTenant(data TenantConfiguration) (TenantConfiguration, error)
	GenerateTenant() (TenantConfiguration, error)
	UpdateTenant(data TenantConfiguration) (TenantConfiguration, error)
	DeleteTenant(identifier []byte) error
}

// configuration holds the configuration details for the node.
type configuration struct {
	mu         sync.RWMutex
	configFile string
	v          *viper.Viper
}

func (c *configuration) Type() reflect.Type {
	panic("irrelevant, configuration#Type must not be used")
}

func (c *configuration) JSON() ([]byte, error) {
	panic("irrelevant, configuration#JSON must not be used")
}

func (c *configuration) FromJSON(json []byte) error {
	panic("irrelevant, configuration#FromJSON must not be used")
}

func (c *configuration) CreateProtobuf() *configpb.ConfigData {
	panic("irrelevant, configuration#CreateProtobuf must not be used")
}

// AccountConfig holds the account details.
type AccountConfig struct {
	Address  string
	Key      string
	Password string
}

// IsSet check if the key is set in the config.
func (c *configuration) IsSet(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.IsSet(key)
}

// Set update the key and the value it holds in the configuration.
func (c *configuration) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v.Set(key, value)
}

// SetDefault sets the default value for the given key.
func (c *configuration) SetDefault(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v.SetDefault(key, value)
}

// Get returns associated value for the key.
func (c *configuration) Get(key string) interface{} {
	return c.get(key)
}

// GetString returns value string associated with key.
func (c *configuration) GetString(key string) string {
	return cast.ToString(c.get(key))
}

// GetInt returns value int associated with key.
func (c *configuration) GetInt(key string) int {
	return cast.ToInt(c.get(key))
}

// GetBool returns value bool associated with key.
func (c *configuration) GetBool(key string) bool {
	return cast.ToBool(c.get(key))
}

// GetDuration returns value duration associated with key.
func (c *configuration) GetDuration(key string) time.Duration {
	return cast.ToDuration(c.get(key))
}

func (c *configuration) get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.Get(key)
}

// GetStoragePath returns the data storage backend.
func (c *configuration) GetStoragePath() string {
	return c.GetString("storage.path")
}

// GetConfigStoragePath returns the config storage backend.
func (c *configuration) GetConfigStoragePath() string {
	return c.GetString("configStorage.path")
}

// GetTenantsKeystore returns the tenants keystore location.
func (c *configuration) GetTenantsKeystore() string {
	return c.GetString("tenants.keystore")
}

// GetP2PPort returns P2P Port.
func (c *configuration) GetP2PPort() int {
	return c.GetInt("p2p.port")
}

// GetP2PExternalIP returns P2P External IP.
func (c *configuration) GetP2PExternalIP() string {
	return c.GetString("p2p.externalIP")
}

// GetP2PConnectionTimeout returns P2P Connect Timeout.
func (c *configuration) GetP2PConnectionTimeout() time.Duration {
	return c.GetDuration("p2p.connectTimeout")
}

// GetReceiveEventNotificationEndpoint returns the webhook endpoint defined in the config.
func (c *configuration) GetReceiveEventNotificationEndpoint() string {
	return c.GetString("notifications.endpoint")
}

// GetServerPort returns the defined server port in the config.
func (c *configuration) GetServerPort() int {
	return c.GetInt("nodePort")
}

// GetServerAddress returns the defined server address of form host:port in the config.
func (c *configuration) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.GetString("nodeHostname"), c.GetString("nodePort"))
}

// GetNumWorkers returns number of queue workers defined in the config.
func (c *configuration) GetNumWorkers() int {
	return c.GetInt("queue.numWorkers")
}

// GetWorkerWaitTimeMS returns the queue worker sleep time between cycles.
func (c *configuration) GetWorkerWaitTimeMS() int {
	return c.GetInt("queue.workerWaitTimeMS")
}

// GetEthereumNodeURL returns the URL of the Ethereum Node.
func (c *configuration) GetEthereumNodeURL() string {
	return c.GetString("ethereum.nodeURL")
}

// GetEthereumContextReadWaitTimeout returns the read duration to pass for context.Deadline.
func (c *configuration) GetEthereumContextReadWaitTimeout() time.Duration {
	return c.GetDuration("ethereum.contextReadWaitTimeout")
}

// GetEthereumContextWaitTimeout returns the commit duration to pass for context.Deadline.
func (c *configuration) GetEthereumContextWaitTimeout() time.Duration {
	return c.GetDuration("ethereum.contextWaitTimeout")
}

// GetEthereumIntervalRetry returns duration to wait between retries.
func (c *configuration) GetEthereumIntervalRetry() time.Duration {
	return c.GetDuration("ethereum.intervalRetry")
}

// GetEthereumMaxRetries returns the max acceptable retries.
func (c *configuration) GetEthereumMaxRetries() int {
	return c.GetInt("ethereum.maxRetries")
}

// GetEthereumGasPrice returns the gas price to use for a ethereum transaction.
func (c *configuration) GetEthereumGasPrice() *big.Int {
	return big.NewInt(cast.ToInt64(c.get("ethereum.gasPrice")))
}

// GetEthereumGasLimit returns the gas limit to use for a ethereum transaction.
func (c *configuration) GetEthereumGasLimit() uint64 {
	return cast.ToUint64(c.get("ethereum.gasLimit"))
}

// GetEthereumDefaultAccountName returns the default account to use for the transaction.
func (c *configuration) GetEthereumDefaultAccountName() string {
	return c.GetString("ethereum.defaultAccountName")
}

// GetEthereumAccount returns the account details associated with the account name.
func (c *configuration) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	k := fmt.Sprintf("ethereum.accounts.%s", accountName)

	if !c.IsSet(k) {
		return nil, errors.New("no account found with account name %s", accountName)
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
func (c *configuration) GetTxPoolAccessEnabled() bool {
	return c.GetBool("ethereum.txPoolAccessEnabled")
}

// GetNetworkString returns defined network the node is connected to.
func (c *configuration) GetNetworkString() string {
	return c.GetString("centrifugeNetwork")
}

// GetNetworkKey returns the specific key(k) value defined in the default network.
func (c *configuration) GetNetworkKey(k string) string {
	return fmt.Sprintf("networks.%s.%s", c.GetNetworkString(), k)
}

// GetContractAddressString returns the deployed contract address for a given contract.
func (c *configuration) GetContractAddressString(contract string) (address string) {
	return c.GetString(c.GetNetworkKey(fmt.Sprintf("contractAddresses.%s", contract)))
}

// GetContractAddress returns the deployed contract address for a given contract.
func (c *configuration) GetContractAddress(contractName ContractName) common.Address {
	return common.HexToAddress(c.GetContractAddressString(string(contractName)))
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *configuration) GetBootstrapPeers() []string {
	return cast.ToStringSlice(c.get(c.GetNetworkKey("bootstrapPeers")))
}

// GetNetworkID returns the numerical network id.
func (c *configuration) GetNetworkID() uint32 {
	return uint32(c.GetInt(c.GetNetworkKey("id")))
}

// GetIdentityID returns the self centID in bytes.
func (c *configuration) GetIdentityID() ([]byte, error) {
	id, err := hexutil.Decode(c.GetString("identityId"))
	if err != nil {
		return nil, centerrors.Wrap(err, "can't read identityId from config")
	}
	return id, err
}

// GetSigningKeyPair returns the signing key pair.
func (c *configuration) GetSigningKeyPair() (pub, priv string) {
	return c.GetString("keys.signing.publicKey"), c.GetString("keys.signing.privateKey")
}

// GetEthAuthKeyPair returns ethereum key pair.
func (c *configuration) GetEthAuthKeyPair() (pub, priv string) {
	return c.GetString("keys.ethauth.publicKey"), c.GetString("keys.ethauth.privateKey")
}

// IsPProfEnabled returns true if the pprof is enabled
func (c *configuration) IsPProfEnabled() bool {
	return c.GetBool("debug.pprof")
}

// LoadConfiguration loads the configuration from the given file.
func LoadConfiguration(configFile string) Configuration {
	cfg := &configuration{configFile: configFile, mu: sync.RWMutex{}}
	cfg.initializeViper()
	return cfg
}

func (c *configuration) readConfigFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	err = c.v.MergeConfig(file)
	return err
}

// initializeViper loads viper if not loaded already.
// This method should not have any effects if Viper is already initialized.
func (c *configuration) initializeViper() {
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

// SmartContractAddresses encapsulates the smart contract addresses ne
type SmartContractAddresses struct {
	IdentityFactoryAddr, IdentityRegistryAddr, AnchorRepositoryAddr, PaymentObligationAddr string
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
	p2pConnectTimeout := args["p2pConnectTimeout"].(string)
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
	v.Set("configStorage.path", targetDataDir+"/db/centrifuge_config_data.leveldb")
	v.Set("tenants.keystore", targetDataDir+"/tenants")
	v.Set("identityId", "")
	v.Set("centrifugeNetwork", network)
	v.Set("nodeHostname", "0.0.0.0")
	v.Set("nodePort", apiPort)
	v.Set("p2p.port", p2pPort)
	if p2pConnectTimeout != "" {
		v.Set("p2p.connectTimeout", p2pConnectTimeout)
	}
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

	if smartContractAddresses, ok := args["smartContractAddresses"].(*SmartContractAddresses); ok {
		v.Set("networks."+network+".contractAddresses.identityFactory", smartContractAddresses.IdentityFactoryAddr)
		v.Set("networks."+network+".contractAddresses.identityRegistry", smartContractAddresses.IdentityRegistryAddr)
		v.Set("networks."+network+".contractAddresses.anchorRepository", smartContractAddresses.AnchorRepositoryAddr)
		v.Set("networks."+network+".contractAddresses.paymentObligation", smartContractAddresses.PaymentObligationAddr)
	}

	v.SetConfigFile(targetDataDir + "/config.yaml")

	err = v.WriteConfig()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return v, nil
}

func (c *configuration) SetupSmartContractAddresses(network string, smartContractAddresses *SmartContractAddresses) {
	c.v.Set("networks."+network+".contractAddresses.identityFactory", smartContractAddresses.IdentityFactoryAddr)
	c.v.Set("networks."+network+".contractAddresses.identityRegistry", smartContractAddresses.IdentityRegistryAddr)
	c.v.Set("networks."+network+".contractAddresses.anchorRepository", smartContractAddresses.AnchorRepositoryAddr)
	c.v.Set("networks."+network+".contractAddresses.paymentObligation", smartContractAddresses.PaymentObligationAddr)
}
